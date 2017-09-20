Installation and Upgrade
================================

pip install
--------------------

::

    pip install toughradius


Source install
------------------------

develop version

::

    pip install -U https://github.com/talkincode/ToughRADIUS/archive/develop.zip

master version

::

    pip install -U https://github.com/talkincode/ToughRADIUS/archive/master.zip


rpm install
-------------------------

open link https://pypi.python.org/pypi/toughradius, copy rpm package link

::

    rpm Uvh {rpm package url}

for example:

::

    rpm Uvh https://pypi.python.org/packages/f9/c0/e2f61b4329239ca2489dc61517086a8c0cabaa0bb59ebcab27b37ae8a05e/toughradius-5.0.0.5-1.noarch.rpm#md5=1b89a6645c5909dbac2ab0fffbc016e5


Upgrade
---------------

develop version

::

    gtrcli upgrade --develop


stable version

::

    gtrcli upgrade --stable