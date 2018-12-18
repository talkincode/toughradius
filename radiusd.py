#!/usr/bin/env python
# coding: utf-8
import sys
sys.path.insert(0, ".")
from toughradius.common import radiusd
if __name__ == "__main__":
    radiusd.run()