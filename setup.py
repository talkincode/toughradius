#!/usr/bin/env python
#coding:utf-8

from setuptools import setup, find_packages
import toughradius
import os

version = toughradius.__version__
proj_home = os.path.dirname(__file__)

install_requires = [
    'gevent==1.2.2',
    'geventhttpclient',
]
install_requires_empty = []

package_data={
    "toughradius" : ["dictionarys/dictionary","dictionarys/dictionary.*"]
}


data_files=[

]


setup(name='toughradius',
      version=version,
      author='jamiesun',
      author_email='jamiesun.net@gmail.com',
      url='https://github.com/talkincode/toughradius',
      license='Apache License 2.0',
      description='Beautiful open source RadiusServer',
      long_description=open('README.md').read(),
      classifiers=[
       'Development Status :: 6 - Mature',
       'Intended Audience :: Developers',
       'Programming Language :: Python :: 2.7',
       'Topic :: Software Development :: Libraries :: Python Modules',
       'Topic :: System :: Systems Administration :: Authentication/Directory',
       ],
      packages=find_packages(),
      package_data=package_data,
      data_files=data_files,
      keywords=['radius', 'AAA','authentication','accounting','authorization','toughradius'],
      zip_safe=True,
      include_package_data=True,
      eager_resources=['toughradius'],
      install_requires=install_requires,
      entry_points={
          'console_scripts': [
              'radiusd = toughradius.common.radiusd:run',
              'authtest = toughradius.common.radtest:test_auth',
              'accttest = toughradius.common.radtest:test_acct',
              'radius-benchmark = toughradius.common.benchmark:start',
          ]
      }
)