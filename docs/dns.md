# DNS Support
If you point a domain or subdomain to Google Cloud DNS as an NS record, kCTF will create a DNS zone that maps the challenge to your challenge's IP address.

Follow these steps:

1. Configure the subdomain in `kctf-setup-config-create`. 

   Example: If you enter *foo.example.com*, the challenges will be deployed on *xxxxx.foo.example.com*.
1. Run `kctf-cluster-start`
1. Point the subdomain to Google Cloud DNS. This requires [adding an NS record on your name server](https://cloud.google.com/dns/docs/update-name-servers).
