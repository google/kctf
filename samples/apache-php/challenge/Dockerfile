FROM kctf-nsjail

RUN ln -s /secrets/flag /chroot/flag

VOLUME /tmp
VOLUME /var/log/apache2
VOLUME /var/run/apache2

RUN apt-get update && apt-get install -yq --no-install-recommends apache2 \
    && rm -rf /var/lib/apt/lists/*
RUN chroot /chroot apt-get update \
    && chroot /chroot apt-get install -yq --no-install-recommends php-cgi \
    && rm -rf /chroot/var/lib/apt/lists/*

RUN service apache2 start

RUN ln -s /etc/apache2/mods-available/cgi.load /etc/apache2/mods-enabled/cgi.load
RUN ln -s /etc/apache2/mods-available/actions.load /etc/apache2/mods-enabled/actions.load

RUN ln -s /config/apache2-nsjail-php.conf /etc/apache2/conf-enabled/nsjail-php.conf

RUN rm -rf /var/www
RUN ln -s /chroot/var/www /var/www

RUN rm -rf /chroot/var/lib/php/sessions
RUN chroot /chroot rm -rf /var/lib/php/sessions
RUN chroot /chroot mkdir -p /mnt/disks/sessions /mnt/disks/uploads
RUN chroot /chroot ln -s /mnt/disks/sessions /var/lib/php/sessions

COPY files/html/ /chroot/var/www/html/
COPY files/cgi-bin/ /usr/lib/cgi-bin/
COPY files/root/ /root/

RUN chmod 0755 /chroot/var/www/html /usr/lib/cgi-bin/
RUN chmod 0755 /chroot/var/www/html/*.php /usr/lib/cgi-bin/nsjail-php-cgi /root/launch-apache

CMD /root/launch-apache
