#coding=utf-8
from toughradius.modules import get_radius_attr
from toughradius.modules import parse_cisco_vlan
from toughradius.modules import VENDOR_RADBACK
import logging

logger = logging.getLogger(__name__)

class RequestParse(object):

    parse_vlan = parse_cisco_vlan

    @staticmethod
    def parse_mac(req):
        mac_addr = get_radius_attr(req, 'Mac-Addr')
        if mac_addr:
            req.client_mac = mac_addr.replace('-', ':')
        return req


    @staticmethod
    def handle_radius(req, debug=False):
        try:
            if int(req.vendor_id) == VENDOR_RADBACK:
                req = RequestParse.parse_mac(req)
                req = RequestParse.parse_vlan(req)
        except Exception as err:
            logger.error("request parse error {}".format(err.message),exc_info=debug)

        return req


class RateLimit(object):

    @staticmethod
    def handle_radius(req, reply, debug=False):
        try:
            if int(req.vendor_id) == VENDOR_RADBACK:
                rate_code = reply.resp_attrs.get('rate_code')
                if rate_code:
                    reply['Sub-Profile-Name'] = str(rate_code)
        except Exception as err:
            logger.error("rate limit error {}".format(err.message), exc_info=debug)

        return reply

request_parse = RequestParse
rate_limit = RateLimit

__all__ = ['request_parse', 'rate_limit']