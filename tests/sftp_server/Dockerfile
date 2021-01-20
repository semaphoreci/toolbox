FROM ubuntu:18.04

RUN apt-get update && apt-get install -y openssh-server
RUN mkdir -p /var/run/sshd

COPY sshd_config /etc/ssh/sshd_config

RUN addgroup ftpaccess
RUN adduser tester --ingroup ftpaccess --shell /usr/sbin/nologin --disabled-password --gecos ''
RUN chown tester:ftpaccess /home/tester
RUN mkdir /home/tester/.ssh

COPY id_rsa.pub id_rsa.pub
RUN cat id_rsa.pub >> /home/tester/.ssh/authorized_keys

EXPOSE 22
CMD ["/usr/sbin/sshd", "-D"]
