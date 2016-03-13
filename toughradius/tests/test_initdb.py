from toughradius.common import initdb
from toughlib import config
from twisted.trial import unittest
import os

class InitdbTestCase(unittest.TestCase):

    def setUp(self):
        os.mkdir("/tmp/toughradius")

    def test_update(self):
        testfile = os.path.join(os.path.abspath(os.path.dirname(__file__)),"test.json")
        iconfig = config.find_config(testfile)
        result = initdb.update(iconfig)
        self.assertEqual(result,None)
