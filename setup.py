#!/usr/bin/python

from setuptools import setup, find_packages
import toughradius
import os

version = toughradius.__version__
proj_home = os.path.dirname(__file__)
configs = os.listdir(os.path.join(proj_home,'etc'))
dictionarys = os.listdir(os.path.join(proj_home,'etc/dictionarys'))

install_requires = [
    'six>=1.8.0',
    'gevent==1.1.2',
    'Click',
    'bottle',
    #'ConcurrentLogHandler'
]
install_requires_empty = []

package_data={}

data_files=[
    ('/etc/toughradius', [ 'etc/%s'%cfg for cfg in configs if cfg not in ('dictionarys',) ]),
    ('/etc/toughradius/dictionarys',['dictionarys/%s'%d for d in dictionarys])
]

setup(name='toughradius',
      version=version,
      author='jamiesun',
      author_email='jamiesun.net@gmail.com',
      url='https://github.com/talkincode/toughradius',
      license='Apache License 2.0',
      description='RADIUS Server',
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
              'gtrad = toughradius.common.commands:cli',
          ]
      }
)