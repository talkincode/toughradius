# -*- coding: utf-8 -*-
from __future__ import unicode_literals

from django.db import models
from django.contrib.auth.models import User,Group
from django.contrib.auth.models import AbstractUser
from django.utils.translation import ugettext_lazy as _
import django.utils.timezone

def get_custom_anon_user(User):
    return User(
        username='AnonymousUser',
    )


class Isp(models.Model):
    ''' Service provider
    '''
    class Meta:
        default_permissions = ()

    code = models.CharField(max_length=32,null=True)
    name = models.CharField(max_length=32,null=True)
    enable = models.BooleanField(default=True)
    remark = models.CharField(max_length=255,null=True)


class Node(models.Model):
    class Meta:
        default_permissions = ()

    isp = models.ForeignKey(Isp,on_delete=models.CASCADE)
    name = models.CharField(max_length=32,null=True)
    enable = models.BooleanField(default=True)
    remark = models.CharField(max_length=255,null=True)


class Operator(AbstractUser):

    class Meta:
        default_permissions = ()

    isp = models.ForeignKey(Isp,on_delete=models.CASCADE)
    nodes = models.ManyToManyField(Node,verbose_name=_('nodes'),blank=True)


class Customer(models.Model):
    ''' ctype 0-normal user 1-group user
    '''
    class Meta:
        default_permissions = ()

    fullname = models.CharField(max_length=128,null=True)
    ctype = models.IntegerField(default=0)
    age = models.IntegerField(default=0)
    sex = models.IntegerField(default=0)
    idcard = models.CharField(max_length=32,null=True)
    mobile = models.CharField(max_length=32,null=True)
    address = models.CharField(max_length=128,null=True)
    wechat = models.CharField(max_length=32,null=True)
    qq = models.CharField(max_length=16,null=True)    
    level = models.CharField(max_length=16,null=True)
    remark = models.CharField(max_length=255,null=True)
    isp = models.ForeignKey(Isp,on_delete=models.CASCADE)
    node = models.ForeignKey(Node,on_delete=models.CASCADE)
    create_time = models.DateTimeField(default=django.utils.timezone.now, verbose_name='create_time')
    update_time = models.DateTimeField(default=django.utils.timezone.now, verbose_name='update_time')


class Wallet(models.Model):

    class Meta:
        default_permissions = ()

    customer = models.ForeignKey(Customer,on_delete=models.CASCADE)
    balance = models.CharField(max_length=32,null=False)
    time_length = models.IntegerField(default=0) 
    flow_length = models.IntegerField(default=0) 
    enable = models.BooleanField(default=True)

class Billuser(models.Model):

    class Meta:
        default_permissions = ()

    customer = models.ForeignKey(Customer,on_delete=models.CASCADE)
    product = models.ForeignKey(Product,on_delete=models.CASCADE)
    username = models.CharField(max_length=64,null=False)
    password = models.CharField(max_length=128,null=False)
    address = models.CharField(max_length=255,null=False)
    enable = models.BooleanField(default=True)
    begin_date = models.DateTimeField(default=django.utils.timezone.now, verbose_name='begin_date')
    expire_date = models.DateTimeField(default=django.utils.timezone.now, verbose_name='expire_date')

class Userattr(models.Model):

    attrtypes = ['radius','normal']

    class Meta:
        default_permissions = ()

    billuser = models.ForeignKey(Billuser,on_delete=models.CASCADE)
    attrtype = models.CharField(max_length=8,null=False)
    name = models.CharField(max_length=64,null=False)
    value = models.CharField(max_length=255,null=False)
    remark = models.CharField(max_length=255,null=True)     


class Billorder(models.Model):

    class Meta:
        default_permissions = ()

    wallet = models.ForeignKey(Wallet,on_delete=models.CASCADE)
    billuser = models.ForeignKey(Billuser,on_delete=models.CASCADE)
    product = models.ForeignKey(Product,on_delete=models.CASCADE)
    order_fee = models.CharField(max_length=32,null=False)
    actual_fee = models.CharField(max_length=32,null=False)
    pay_done = models.BooleanField(default=False) 
    check_done = models.BooleanField(default=False) 
    remark = models.CharField(max_length=255,null=True)
    create_time = models.DateTimeField(default=django.utils.timezone.now, verbose_name='create_time')
    update_time = models.DateTimeField(default=django.utils.timezone.now, verbose_name='update_time')


class Nas(models.Model):

    class Meta:
        default_permissions = ()

    ipv4_addr = models.CharField(max_length=32,null=False)
    ipv6_addr = models.CharField(max_length=32,null=True)
    auth_port = models.IntegerField(default=1812)
    acct_port = models.IntegerField(default=1813)
    secret = models.CharField(max_length=32,null=False)
    timezone = models.CharField(max_length=32,null=True)
    enable = models.BooleanField(default=True)
    isps = models.ManyToManyField(Isp,verbose_name=_('isps'),blank=True)


class Period(models.Model):

    class Meta:
        default_permissions = ()

    start_time = models.CharField(max_length=8,null=False)
    end_time = models.CharField(max_length=8,null=False)



class Service(models.Model):

    class Meta:
        default_permissions = ()

    name = models.CharField(max_length=64,null=False)
    limit = models.IntegerField(default=1)
    bind_mac = models.BooleanField(default=False)
    bind_vlan = models.BooleanField(default=False)
    input_limit = models.IntegerField(default=0)
    output_limit = models.IntegerField(default=0)
    fee_days = models.IntegerField(default=0)
    fee_months = models.IntegerField(default=0)
    fee_times = models.IntegerField(default=0)
    fee_flows = models.IntegerField(default=0)
    max_price = models.CharField(max_length=32,null=False)
    min_price = models.CharField(max_length=32,null=False)
    enable = models.BooleanField(default=True)
    remark = models.CharField(max_length=255,null=True)
    periods = models.ManyToManyField(Period,verbose_name=_('periods'),blank=True)


class Product(models.Model):

    class Meta:
        default_permissions = ()

    name = models.CharField(max_length=128,null=False)
    service = models.ForeignKey(Service,on_delete=models.DO_NOTHING)
    price = models.CharField(max_length=32,null=False)
    ispub = models.BooleanField(default=False)
    enable = models.BooleanField(default=True)
    sale_expire = models.DateTimeField(default=django.utils.timezone.now, verbose_name='sale_expire')
    create_time = models.DateTimeField(default=django.utils.timezone.now, verbose_name='create_time')
    update_time = models.DateTimeField(default=django.utils.timezone.now, verbose_name='update_time')
    isp = models.ForeignKey(Isp,on_delete=models.CASCADE)









