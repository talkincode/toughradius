#!/usr/bin/env python
# -*- coding: utf-8 -*-

import os
import sys

__version__ = '0.0.1'

try:
    from setuptools import setup,find_packages
except ImportError:
    from distutils.core import setup,find_packages


setup(
    name='ToughRADIUS',
    version=__version__,
    description='ToughRADIUS is a Radius Server',
    long_description=(open('README.md').read()),
    author='jamiesun',
    author_email='jamiesun.net@gmail.com',
    url='http://www.toughradius.org',
    packages=find_packages(exclude=['tests','views','static','dict','build']),
    include_package_data=True,
    license='MIT',
    classifiers=(
        'Development Status :: 5 - Production/Stable',
        'Intended Audience :: Developers',
        'Natural Language :: English',
        'License :: OSI Approved :: MIT License',
        'Programming Language :: Python',
        'Programming Language :: Python :: 2.5',
        'Programming Language :: Python :: 2.6',
        'Programming Language :: Python :: 2.7',
    ),
    keywords=['radius', 'aaa'],
    # data_files=[
    #     ('radiusd/dict', ['radiusd/dict/*']),
    #     ('console/views', ['console/views/*']),
    #     ('console/static/css', ['console/static/css/*']),
    #     ('console/static/fonts', ['console/static/fonts/*']),
    #     ('console/static/js', ['console/static/js/*']),
    #     ('console/static/img', ['console/static/img/*']),
    #     ('console/static', ['console/static/*.ico']),
    # ]    
)