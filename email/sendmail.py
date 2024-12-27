#!/usr/bin/python3

from email.message import EmailMessage
from pathlib import Path
from os import environ
from smtplib import SMTP_SSL
from sys import argv, exit, stderr


RC_OK = 0
RC_ERR = 1

SCRIPT_DIR = Path(__file__).parent


def main(argv):
    smtp_user = environ.get("SMTP_USERNAME")
    smtp_pass = environ.get("SMTP_PASSWORD")
    smtp_host = environ.get("SMTP_HOSTNAME")
    smtp_port = environ.get("SMTP_PORT")
    mail_from = environ.get("MAIL_FROM")

    assert isinstance(smtp_host, str)
    assert isinstance(smtp_port, str)
    assert isinstance(smtp_user, str)
    assert isinstance(smtp_pass, str)
    assert isinstance(mail_from, str)

    email_tmpl_path_s, mail_to_address, mail_to_name = argv
    email_tmpl_path = Path(email_tmpl_path_s)

    if not "@" in mail_to_address:
        print("first argument is not an email address", file=stderr)
        return RC_ERR

    mail_to = f"{ mail_to_name } <{ mail_to_address }>"
    email = make_email(mail_from, mail_to, email_tmpl_path)

    s = SMTP_SSL(host=smtp_host, port=int(smtp_port))
    s.login(smtp_user, smtp_pass)
    s.sendmail(mail_from, [mail_to], email.as_string())

    return RC_OK



def make_email(mail_from, mail_to, email_tmpl_path):
    email = EmailMessage()
    email.make_alternative()
    email['From'] = mail_from
    email['To'] = mail_to

    with open(email_tmpl_path / "subject", "r") as f:
        subject = f.read().strip("\n")
    email['Subject'] = subject

    with open(email_tmpl_path / "email.txt", "r") as f:
        plain = f.read().rstrip("\n")
    email.add_alternative(plain, subtype="plain")

    with open(email_tmpl_path / "email.html", "r") as f:
        html = f.read().rstrip("\n")
    email.add_alternative(html, subtype="html")

    return email

if __name__ == "__main__":
    exit(main(argv[1:]))
