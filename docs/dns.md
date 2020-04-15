# DNS Support
If you point a domain or subdomain to Google Cloud DNS, you can use `kctf-setup-dns` and `kctf-chal-dns` to map *[challenge_name].yourdomain.com* to your challenge.

Follow these steps:

1. Configure the subdomain in `kctf-setup-config-create`. 

   Example: If you enter *foo.example.com*, the challenges will be deployed on *[challenge-name].foo.example.com*.
1. Run `kctf-setup-dns` to create the DNS zone in Google Cloud DNS.
1. Point the subdomain to Google Cloud DNS. This requires [adding an NS record on your name server](https://cloud.google.com/dns/docs/update-name-servers).
1. Run `kctf-chal-expose [challenge_name]` to expose your challenge.
1. Run `kctf-chal-dns [challenge_name]` to set up the subdomain.
