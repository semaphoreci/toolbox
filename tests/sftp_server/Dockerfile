FROM ubuntu:18.04

RUN apt-get update && apt-get install -y openssh-server
RUN mkdir -p /var/run/sshd

COPY sshd_config /etc/ssh/sshd_config

RUN addgroup ftpaccess
RUN adduser tester --ingroup ftpaccess --shell /bin/bash --disabled-password --gecos ''
RUN chown tester:ftpaccess /home/tester
RUN mkdir /etc/ssh/authorized_keys

COPY id_rsa.pub /tmp/id_rsa.pub
RUN cat /tmp/id_rsa.pub >> /etc/ssh/authorized_keys/tester

EXPOSE 22
CMD ["/usr/sbin/sshd", "-D"]
