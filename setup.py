
#!/usr/bin/python

from setuptools import setup, find_packages
import toughradius

version = toughradius.__version__

install_requires = [
    'six>=1.8.0',
    'gevent==1.1.2',
    'Click',
    'bottle'
]
install_requires_empty = []

package_data={}

data_files=[
    ('/etc/toughradius', [
        'etc/radiusd.json', 
        'etc/logger.json'
        'etc/clients.json'
        'etc/modules.json'
    ]),
    ('/etc/toughradius/dictionarys',['dictionarys/directory','dictionarys/directory.*'])
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
      keywords=['radius', 'AAA','authentication','accounting','authorization','toughradius'],
      zip_safe=True,
      include_package_data=True,
      eager_resources=['toughradius'],
      install_requires=install_requires,
      entry_points={
          'console_scripts': [
              'gtr-auth = toughradius.radiusd.server:auth',
              'gtr-acct = toughradius.radiusd.server:acct',
          ]
      }
)