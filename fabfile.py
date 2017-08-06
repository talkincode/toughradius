#!/usr/bin/env python
import sys,os
sys.path.insert(0,os.path.dirname(__file__))
from fabric.api import *
# from fabric.contrib.project import rsync_project
import datetime

def rundev():
    local("python manage.py runserver 0.0.0.0:8000")

