#!/usr/bin/env python
#coding=utf-8

from django.conf.urls import url
from django.contrib import admin

urlpatterns = [
    url(r'^admin/', admin.site.urls),
]
