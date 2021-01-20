sudo sh -c 'echo "Subsystem sftp internal-sftp" >> /private/etc/ssh/sshd_config"'

sudo launchctl stop com.openssh.sshd
sudo launchctl start com.openssh.sshd
