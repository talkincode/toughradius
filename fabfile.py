#!/usr/bin/env python
import sys,os
sys.path.insert(0,os.path.dirname(__file__))
from fabric.api import *
# from fabric.contrib.project import rsync_project
import datetime

def rundev():
    local("python manage.py runserver 0.0.0.0:8000")

def upcore():
    local("python manage.py makemigrations core")

def syncdb():
    local("python manage.py migrate")

def adduser():
    local("python manage.py createsuperuser")

def mkmsg():
    local("django-admin makemessages -l zh")

def compmsg():
    local("django-admin compilemessages")

def loaddata():
    local("python manage.py loaddata fixtures/winapps")

def dbsh():
    local("python manage.py dbshell")