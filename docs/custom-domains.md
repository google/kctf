# Custom Domains

When creating your cluster, you can specify a domain with the `--domain-name` flag.
kCTF will then automatically create domain names for challenges of the form:
* $chal\_name.$kctf\_domain for TCP based challenges
* $chal\_name-web.$kctf\_domain for HTTPS based challenges

You might want to use custom domains for some of your challenges, for example:
* if you need to have a challenge available on multiple host names
* to protect web challenges against same-site attacks
* or simply if you want to have a fancy domain name

For TCP based challenges, all you need to do is to create a CNAME DNS entry from $cooldomain to $chal\_name.$kctf\_domain.

For HTTPS based challenges, you also need to add a CNAME entry (pay attention to the -web suffix) and in addition, list the domain in the port configuration of the challenge:
```yaml
apiVersion: kctf.dev/v1
kind: Challenge
metadata:
  name: web
spec:
  deployed: true
  powDifficultySeconds: 0
  network:
    public: true
    ports:
      - protocol: "HTTPS"
        targetPort: 1337
        domains:
          - "cooldomain.com"
```
With this, kCTF will automatically create a certificate for you and attach it to the challenge's LoadBalancer.
