import unittest
import six
from pyrad import tools


class EncodingTests(unittest.TestCase):
    def testStringEncoding(self):
        self.assertRaises(ValueError, tools.EncodeString, 'x' * 254)
        self.assertEqual(
                tools.EncodeString('1234567890'),
                six.b('1234567890'))

    def testInvalidStringEncodingRaisesTypeError(self):
        self.assertRaises(TypeError, tools.EncodeString, 1)

    def testAddressEncoding(self):
        self.assertRaises(ValueError, tools.EncodeAddress, '123')
        self.assertEqual(
                tools.EncodeAddress('192.168.0.255'),
                six.b('\xc0\xa8\x00\xff'))

    def testInvalidAddressEncodingRaisesTypeError(self):
        self.assertRaises(TypeError, tools.EncodeAddress, 1)

    def testIntegerEncoding(self):
        self.assertEqual(tools.EncodeInteger(0x01020304),
                six.b('\x01\x02\x03\x04'))

    def testUnsignedIntegerEncoding(self):
        self.assertEqual(tools.EncodeInteger(0xFFFFFFFF),
                six.b('\xff\xff\xff\xff'))

    def testInvalidIntegerEncodingRaisesTypeError(self):
        self.assertRaises(TypeError, tools.EncodeInteger, '1')

    def testDateEncoding(self):
        self.assertEqual(tools.EncodeDate(0x01020304),
                six.b('\x01\x02\x03\x04'))

    def testInvalidDataEncodingRaisesTypeError(self):
        self.assertRaises(TypeError, tools.EncodeDate, '1')

    def testStringDecoding(self):
        self.assertEqual(
                tools.DecodeString(six.b('1234567890')),
                '1234567890')

    def testAddressDecoding(self):
        self.assertEqual(
                tools.DecodeAddress(six.b('\xc0\xa8\x00\xff')),
                '192.168.0.255')

    def testIntegerDecoding(self):
        self.assertEqual(
                tools.DecodeInteger(six.b('\x01\x02\x03\x04')),
                0x01020304)

    def testDateDecoding(self):
        self.assertEqual(
                tools.DecodeDate(six.b('\x01\x02\x03\x04')),
                0x01020304)

    def testUnknownTypeEncoding(self):
        self.assertRaises(ValueError, tools.EncodeAttr, 'unknown', None)

    def testUnknownTypeDecoding(self):
        self.assertRaises(ValueError, tools.DecodeAttr, 'unknown', None)

    def testEncodeFunction(self):
        self.assertEqual(
                tools.EncodeAttr('string', six.u('string')),
                six.b('string'))
        self.assertEqual(
                tools.EncodeAttr('octets', six.b('string')),
                six.b('string'))
        self.assertEqual(
                tools.EncodeAttr('ipaddr', '192.168.0.255'),
                six.b('\xc0\xa8\x00\xff'))
        self.assertEqual(
                tools.EncodeAttr('integer', 0x01020304),
                six.b('\x01\x02\x03\x04'))
        self.assertEqual(
                tools.EncodeAttr('date', 0x01020304),
                six.b('\x01\x02\x03\x04'))

    def testDecodeFunction(self):
        self.assertEqual(
                tools.DecodeAttr('string', six.b('string')),
                six.u('string'))
        self.assertEqual(
                tools.EncodeAttr('octets', six.b('string')),
                six.b('string'))
        self.assertEqual(
                tools.DecodeAttr('ipaddr', six.b('\xc0\xa8\x00\xff')),
                '192.168.0.255')
        self.assertEqual(
                tools.DecodeAttr('integer', six.b('\x01\x02\x03\x04')),
                0x01020304)
        self.assertEqual(
                tools.DecodeAttr('date', six.b('\x01\x02\x03\x04')),
                0x01020304)
