package enmime

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
	"strings"
)

// MIMEPart is the primary interface enmine clients will use.  Each MIMEPart represents
// a node in the MIME multipart tree.  The Content-Type, Disposition and File Name are
// parsed out of the header for easier access.
//
// TODO Content should probably be a reader so that it does not need to be stored in
// memory.
type MIMEPart interface {
	Parent() MIMEPart               // Parent of this part (can be nil)
	SetParent(MIMEPart)             // Sets Parent of this part
	FirstChild() MIMEPart           // First (top most) child of this part
	SetFirstChild(MIMEPart)         // Sets first child of this part (we are parent)
	NextSibling() MIMEPart          // Next sibling of this part
	SetNextSibling(MIMEPart)        // Sets next sibling of this part (shares our parent)
	Header() textproto.MIMEHeader   // Header as parsed by textproto package
	SetHeader(textproto.MIMEHeader) // Sets the part MIME header
	ContentType() string            // Content-Type header without parameters
	SetContentType(string)          // Sets the Content-Type header
	Disposition() string            // Content-Disposition header without parameters
	SetDisposition(string)          // Sets the Content-Disposition header
	FileName() string               // File Name from disposition or type header
	SetFileName(string)             // Set file name
	Charset() string                // Content character set label
	SetCharset(string)              // Set content character set label
	Content() []byte                // Decoded content of this part (can be empty)
	SetContent([]byte)              // Set content of this part
}

// memMIMEPart is an in-memory implementation of the MIMEPart interface.  It will likely
// choke on huge attachments.
type memMIMEPart struct {
	parent      MIMEPart
	firstChild  MIMEPart
	nextSibling MIMEPart
	header      textproto.MIMEHeader
	contentType string
	disposition string
	fileName    string
	charset     string
	content     []byte
}

// NewMIMEPart creates a new memMIMEPart object.  It does not update the parents FirstChild
// attribute.
func NewMIMEPart(parent MIMEPart, contentType string) MIMEPart {
	return &memMIMEPart{parent: parent, contentType: contentType}
}

// Parent of this part (can be nil)
func (p *memMIMEPart) Parent() MIMEPart {
	return p.parent
}

// SetParent sets the parent of this part
func (p *memMIMEPart) SetParent(parent MIMEPart) {
	p.parent = parent
}

// First (top most) child of this part
func (p *memMIMEPart) FirstChild() MIMEPart {
	return p.firstChild
}

// SetFirstChild sets the first (top most) child of this part
func (p *memMIMEPart) SetFirstChild(child MIMEPart) {
	p.firstChild = child
}

// NextSibling of this part
func (p *memMIMEPart) NextSibling() MIMEPart {
	return p.nextSibling
}

// SetNextSibling sets the next sibling (shares parent) of this part
func (p *memMIMEPart) SetNextSibling(sibling MIMEPart) {
	p.nextSibling = sibling
}

// Header as parsed by textproto package
func (p *memMIMEPart) Header() textproto.MIMEHeader {
	return p.header
}

// SetHeader sets a MIME part header.
func (p *memMIMEPart) SetHeader(header textproto.MIMEHeader) {
	p.header = header
}

// Content-Type header without parameters
func (p *memMIMEPart) ContentType() string {
	return p.contentType
}

// SetContentType sets the Content-Type.
// Example: "image/jpg" or "application/octet-stream"
func (p *memMIMEPart) SetContentType(contentType string) {
	p.contentType = contentType
}

// Dispostion returns the Content-Disposition header without parameters
func (p *memMIMEPart) Disposition() string {
	return p.disposition
}

// SetDisposition sets the Content-Disposition.
// Example: "attachment" or "inline"
func (p *memMIMEPart) SetDisposition(disposition string) {
	p.disposition = disposition
}

// FileName returns the file name from disposition or type header
func (p *memMIMEPart) FileName() string {
	return p.fileName
}

// SetFileName sets the parts file name.
func (p *memMIMEPart) SetFileName(fileName string) {
	p.fileName = fileName
}

// Charset returns the content charset encoding label
func (p *memMIMEPart) Charset() string {
	return p.charset
}

// SetCharset sets the charset encoding label of this content, see charsets.go
// for examples, but you probably want "utf-8"
func (p *memMIMEPart) SetCharset(charset string) {
	p.charset = charset
}

// Content returns the decoded content of this part (can be empty)
func (p *memMIMEPart) Content() []byte {
	return p.content
}

func (p *memMIMEPart) SetContent(content []byte) {
	p.content = content
}

// ParseMIME reads a MIME document from the provided reader and parses it into
// tree of MIMEPart objects.
func ParseMIME(reader *bufio.Reader) (MIMEPart, error) {
	tr := textproto.NewReader(reader)
	header, err := tr.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	mediatype, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	root := &memMIMEPart{header: header, contentType: mediatype}

	if strings.HasPrefix(mediatype, "multipart/") {
		boundary := params["boundary"]
		err = parseParts(root, reader, boundary)
		if err != nil {
			return nil, err
		}
	} else {
		// Content is text or data, decode it
		content, err := decodeSection(header.Get("Content-Transfer-Encoding"), reader)
		if err != nil {
			return nil, err
		}
		root.content = content
	}

	return root, nil
}

// parseParts recursively parses a mime multipart document.
func parseParts(parent MIMEPart, reader io.Reader, boundary string) error {
	var prevSibling MIMEPart

	// Loop over MIME parts
	mr := multipart.NewReader(reader, boundary)
	for {
		// mrp is golang's built in mime-part
		mrp, err := mr.NextPart()
		if err != nil {
			if err == io.EOF {
				// This is a clean end-of-message signal
				break
			}
			return err
		}
		if len(mrp.Header) == 0 {
			// Empty header probably means the part didn't using the correct trailing "--"
			// syntax to close its boundary.  We will let this slide if this this the
			// last MIME part.
			if _, err := mr.NextPart(); err != nil {
				if err == io.EOF || strings.HasSuffix(err.Error(), "EOF") {
					// This is what we were hoping for
					break
				} else {
					return fmt.Errorf("Error at boundary %v: %v", boundary, err)
				}
			}

			return fmt.Errorf("Empty header at boundary %v", boundary)
		}
		ctype := mrp.Header.Get("Content-Type")
		if ctype == "" {
			return fmt.Errorf("Missing Content-Type at boundary %v", boundary)
		}
		mediatype, mparams, err := mime.ParseMediaType(ctype)
		if err != nil {
			return err
		}

		// Insert ourselves into tree, p is enmime's MIME part
		p := NewMIMEPart(parent, mediatype)
		p.SetHeader(mrp.Header)
		if prevSibling != nil {
			prevSibling.SetNextSibling(p)
		} else {
			parent.SetFirstChild(p)
		}
		prevSibling = p

		// Determine content disposition, filename, character set
		disposition, dparams, err := mime.ParseMediaType(mrp.Header.Get("Content-Disposition"))
		if err == nil {
			// Disposition is optional
			p.SetDisposition(disposition)
			p.SetFileName(DecodeHeader(dparams["filename"]))
		}
		if p.FileName() == "" && mparams["name"] != "" {
			p.SetFileName(DecodeHeader(mparams["name"]))
		}
		if p.FileName() == "" && mparams["file"] != "" {
			p.SetFileName(DecodeHeader(mparams["file"]))
		}
		if p.Charset() == "" {
			p.SetCharset(mparams["charset"])
		}

		boundary := mparams["boundary"]
		if boundary != "" {
			// Content is another multipart
			err = parseParts(p, mrp, boundary)
			if err != nil {
				return err
			}
		} else {
			// Content is text or data, decode it
			data, err := decodeSection(mrp.Header.Get("Content-Transfer-Encoding"), mrp)
			if err != nil {
				return err
			}
			p.SetContent(data)
		}
	}

	return nil
}

// decodeSection attempts to decode the data from reader using the algorithm listed in
// the Content-Transfer-Encoding header, returning the raw data if it does not known
// the encoding type.
func decodeSection(encoding string, reader io.Reader) ([]byte, error) {
	// Default is to just read input into bytes
	decoder := reader

	switch strings.ToLower(encoding) {
	case "quoted-printable":
		decoder = quotedprintable.NewReader(reader)
	case "base64":
		cleaner := NewBase64Cleaner(reader)
		decoder = base64.NewDecoder(base64.StdEncoding, cleaner)
	}

	// Read bytes into buffer
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(decoder)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
