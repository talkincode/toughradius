#coding=utf-8
from toughradius.modules import parse_cisco_vlan
from toughradius.modules import parse_std_mac
from toughradius.modules import VENDOR_JUNIPER
import logging

logger = logging.getLogger(__name__)

class RequestParse(object):

    parse_vlan = parse_cisco_vlan
    parse_mac = parse_std_mac

    @staticmethod
    def handle_radius(req, debug=False):
        try:
            if int(req.vendor_id) == VENDOR_JUNIPER:
                req = RequestParse.parse_mac(req)
                req = RequestParse.parse_vlan(req)
        except Exception as err:
            logger.error("request parse error {}".format(err.message),exc_info=debug)

        return req

request_parse = RequestParse