#coding=utf-8
from toughradius.modules import parse_cisco_vlan
from toughradius.modules import parse_std_mac
from toughradius.modules import VENDOR_MIKROTIC
import logging

logger = logging.getLogger(__name__)

class RequestParse(object):

    parse_vlan = parse_cisco_vlan
    parse_mac = parse_std_mac

    @staticmethod
    def handle_radius(req, debug=False):
        try:
            if int(req.vendor_id) == VENDOR_MIKROTIC:
                req = RequestParse.parse_mac(req)
                req = RequestParse.parse_vlan(req)
        except Exception as err:
            logger.error("request parse error {}".format(err.message),exc_info=debug)

        return req

class RateLimit(object):

    @staticmethod
    def handle_radius(req, reply, debug=False):
        try:
            if int(req.vendor_id) == VENDOR_MIKROTIC:
                input_rate = int(reply.resp_attrs.get('input_rate', 0)) /1024
                output_rate = int(reply.resp_attrs.get('output_rate', 0))/1024
                reply['Mikrotik-Rate-Limit'] = '%sk/%sk' % (input_rate, output_rate)
        except Exception as err:
            logger.error("rate limit error {}".format(err.message), exc_info=debug)

        return reply



request_parse = RequestParse
rate_limit = RateLimit