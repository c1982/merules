# merules - MTA Pickup Event for MailEnable

Bu araç MailEnable SMTP servisine gelen (Inbound) epostaları belirli kurallar çerçevesinde kontrol eder, kurallara aykırı olan epostaların içeriği temizlenir ve kullanıcıya neden bu epostayı alamadığına dair bilgi geçilir.

## Download

* [https://github.com/c1982/merules/releases](https://github.com/c1982/merules/releases)

## Kurulum

**1)** merules.zip dosyasını açın ve

`C:\Program Files (x86)\Mail Enable\Bin64`

dizinine kopyalayın.

**2)** MailEnable MMC konsolundan 

- Servers > Services and Connectors > MTA 


menüsüne sağ tıklayarak **Properties** ekranına ulaşın.  **Enable Pickup event** kutusunu işaretleyip, **Program to execute on mail file**: alanına merules.exe dosyasını tanıtın.

Daha kolay bir yol ise, aşağıdaki Registry kaytını merules.reg olarak
sunucu üzerine kaydedin ve çalıştırın.

> 
> merulles.reg
> 
> Windows Registry Editor Version 5.00
> 
> [HKEY_LOCAL_MACHINE\SOFTWARE\Wow6432Node\Mail Enable\Mail
> Enable\Agents\MTA]
> 
> "Pickup Event Enabled"=dword:00000001 "Pickup Event
> Command"="C:\\Program Files (x86)\\Mail Enable\\Bin64\\merules.exe"

Bilgi: [https://www.mailenable.com/documentation/6.0/Enterprise/MTA_-_General.html](https://www.mailenable.com/documentation/6.0/Enterprise/MTA_-_General.html)

**3)** merules.config dosyasını kendi ihtiyacınıza göre ayarlayın.

Varsayılan _merules.config_

```toml
MaxScanSizeKB=140
BlockPassZip=true
BlockPassZip_Msg = "Email is cleaned!\nZip files cannot be encrypted like ramsonware: %1\n\nSubject: %2"
BlockExtensions=["exe","msi","bat"]
BlockExtensions_Msg = "Email is cleaned!\nThis mail has contains blocked attachment. Detected file is: %1\n\nSubject: %2"
ScanMalwareDomain=true
ScanMalwareDomain_Msg="Email is cleaned!\nMalware domain detected in email body: %1\n\nSubject: %2"
EmailFooter="\n\n--\nYour mail changed by mail server - merules v1.0"
MailEnablePath="C:\\Program Files (x86)\\Mail Enable"
ScanServices =["SMTP","LS","SF","POP"]
DeleteDetectedMail=false
SendReportRecipient=false
SendReportSender=true
SenderEmail="postmaster@domain.com"
```

**Konfigürasyon Dosyası Açıklamaları**

MaxScanSizeKB

Dikkate alınacak eposta eklerinin (Attachment) maksimum boyutunu belirler. KB cinsinden girilen değeri geçen ekler dikkate alınmaz. Varsayılan değer 140KB'dır. 140KB'ı geçen dosya ekleri dikkate alınmaz.

BlockPassZip

Zip dosyalarının parola korumalı (Encrypted Zip) olup olmadığını kontrol edilmesini sağlar.

BlockPassZip_Msg 

Parola korumalı Zip bulunduğunda kullanıcıya iletilecek mesajı belirler. %1 dosyanın ismi, %2 epostanın konusu.

BlockExtensions

Eposta eklerini yasaklar. Değeri [] olarak girildiğinde dikkate alınmaz.

BlockExtensions_Msg 

Eposta ekleri yasaklandığında kullanıcıya iletilecek mesajı belirler. %1 dosyanın ismi, %2 epostanın konusu.

ScanMalwareDomain

Eposta içeriğinde yasaklı domainlerin aranmasını sağlar.

ScanMalwareDomain_Msg

Yasaklı domain bulunduğunda kullanıcıya iletilecek mesajı belirler. %1 domain'in ismi, %2 epostanın konusu.

EmailFooter

Kullanıcıya iletilecek mesajların altında imza satırını belirler.

MailEnablePath

Sunucu üzerinde MailEnable yazılımının hangi dizinde çalıştığını belirler.

ScanServices

MailEnable servislerinden hangisinin dikkate alınacağını belirler. SMTP dışarıya giden epostaları, SF içeriden dolaşan epostaları, LS liste özelliğinde kullanılan epostaları tarar. Sadece gelenleri taramak için SMTP. Giden ve Gelen epostaları taramak için ise SMTP, SF eğerlerini girmelisiniz.

DeleteDetectedMail

Taranan ve kuralların yakaladığı zararlı epostayı anında siler ve herhangi bir rapor göndermez.

SendReportRecipient

Epostada zararlı yakalandığı zaman alıcıya "Received Failure" raporu gönderir. NDR raporuda denir.

SendReportSender

Epostada zararlı yakalandığı zaman gönderene "Delivery Failure" raporu gönderir. NDR raporuda denir.

SenderEmail

Eposta raporu gönderilirken hangi eposta üzerinden gönderileceğini belirtebilirsiniz. Sunucuda var olan bir eposta adresi belirtmelisiniz.

EnableWhiteList

Gelen email adresinin domain'ini belirledikten sonra whitelist.config içindeki domainlerle karşılaştırır. Eşetiği taktirde mevcut kurallar uygulanmaz.

## Uyguladığı Kurallar

Bu araç ön tanımlı kurallar çevçevesinde hareket eder. Bu kurallar sırası ile şöyledir.

**Yasaklanmış Eposta Eklerinin Bloklanması**

Konfigürasyon dosyasında *BlockExtensions* alanından yönetilebilir. Araç, burada belirtilen dosya uzantılarını epostanın eklerinde arar ve bulduğunda epostanın içeriğine uygun mesajı yazarak alıcısına iletir.

**Yasaklanmış Domain'lerin Bloklanması**

Konfigürasyon dosyasında *ScanMalwareDomain* alanından yönetilebilir. blacklist.config dosyasında belirtilen domainleri eposta'nın BODY kısmında arar. Eğer bu domainlerden biri eposta içinde geçiyorsa epostanın içeriğine uygun mesajı yazarak alıcısına iletir.

> **blacklist.config** dosyasının oluşturulmasında *NormShield.com Suspicious Domain List* servisi kullanılmıştır. 
> Detaylı bilgi için [ https://reputation.normshield.com/](https://reputation.normshield.com/)

**Parola Korumalı Zip Dosyalarının Bloklanması**

Konfigürasyon dosyasında BlockPassZip alanından yönetilir.  Cryptolocker virüsleri genelde Encrypted zip dosyaları ile yayılırlar (yada benim karşılaştığım o yönde). Araç, epostanın zip içerikli eklerini kontrol eder ve zip dosyası encrypted ise epostanın içeriğine uygun mesajı yazarak alıcısına iletir.

## Debug

MailEnable MTA Debug için [https://www.mailenable.com/kb/content/article.asp?ID=ME020121](https://www.mailenable.com/kb/content/article.asp?ID=ME020121) adresini ziyaret ediniz.

## Paketler

* [github.com/jhillyerd/go.enmime](https://github.com/jhillyerd/go.enmime)
* [github.com/alexmullins/zip](https://github.com/alexmullins/zip)
* [github.com/mohamedattahri/mail](https://github.com/mohamedattahri/mail)
* [github.com/BurntSushi/toml](https://github.com/BurntSushi/toml)

## İletişim

Oğuzhan YILMAZ

aspsrc@gmail.com 