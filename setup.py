#!/usr/bin/python

from setuptools import setup, find_packages
import toughradius

version = toughradius.__version__

install_requires = [
    'argparse',
    'Mako>=0.9.0',
    'Beaker>=1.6.4',
    'MarkupSafe>=0.18',
    'PyYAML>=3.10',
    'SQLAlchemy>=0.9.8',
    'Twisted>=13.0.0',
    'autobahn>=0.9.3-3',
    'bottle>=0.12.7',
    'six>=1.8.0',
    'tablib>=0.10.0',
    'zope.interface>=4.1.1',
    'pycrypto==2.6.1'
]
install_requires_empty = []

package_data={
    'toughradius': [
        'console/admin/views/*',
        'console/customer/views/*',
        'console/static/css/*',
        'console/static/fonts/*',
        'console/static/img/*',
        'console/static/js/*',
        'console/static/favicon.ico',
        'radiusd/dicts/*'
    ]
}


setup(name='toughradius',
      version=version,
      author='jamiesun',
      author_email='jamiesun.net@gmail.com',
      url='https://github.com/toughstruct/ToughRADIUS',
      license='GPL',
      description='RADIUS Server',
      long_description=open('README.rst').read(),
      classifiers=[
       'Development Status :: 6 - Mature',
       'Intended Audience :: Developers',
       'License :: OSI Approved :: GPL',
       'Programming Language :: Python :: 2.6',
       'Programming Language :: Python :: 2.7',
       'Topic :: Software Development :: Libraries :: Python Modules',
       'Topic :: System :: Systems Administration :: Authentication/Directory',
       ],
      packages=find_packages(),
      package_data=package_data,
      keywords=['radius', 'authentication'],
      zip_safe=True,
      include_package_data=True,
      install_requires=install_requires,
      scripts=['toughctl'],
      tests_require='nose>=0.10.0b1',
      test_suite='nose.collector',
)