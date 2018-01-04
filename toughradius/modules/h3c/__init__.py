# coding=utf-8

from toughradius.modules import get_radius_attr
from toughradius.modules import parse_std_vlan
from toughradius.modules import VENDOR_H3C
import logging
logger = logging.getLogger(__name__)


class RequestParse(object):
    parse_vlan = parse_std_vlan

    @staticmethod
    def parse_mac(req):
        mac_addr = get_radius_attr(req, 'H3C-Ip-Host-Addr')
        if mac_addr and len(mac_addr) > 17:
            req.client_mac = mac_addr[:-17]
        else:
            req.client_mac = mac_addr

        return req

    @staticmethod
    def handle_radius(req, debug=False):
        try:
            if int(req.vendor_id) == VENDOR_H3C:
                req = RequestParse.parse_mac(req)
                req = RequestParse.parse_vlan(req)
        except Exception as err:
            logger.error("request parse error {}".format(err.message), exc_info=debug)

        return req



class RateLimit(object):

    @staticmethod
    def handle_radius(req, reply, debug=False):
        try:
            if int(req.vendor_id) == VENDOR_H3C:
                input_rate = int(reply.resp_attrs.get('input_rate', 0))
                output_rate = int(reply.resp_attrs.get('output_rate', 0))
                reply['H3C-Input-Average-Rate'] = input_rate
                reply['H3C-Input-Peak-Rate'] = input_rate
                reply['H3C-Output-Average-Rate'] = output_rate
                reply['H3C-Output-Peak-Rate'] = output_rate
        except Exception as err:
            logger.error("rate limit error {}".format(err.message), exc_info=debug)

        return reply

class KbpsRateLimit(object):

    @staticmethod
    def handle_radius(req, reply, debug=False):
        try:
            if int(req.vendor_id) == VENDOR_H3C:
                input_rate = int(reply.resp_attrs.get('input_rate', 0))/1024
                output_rate = int(reply.resp_attrs.get('output_rate', 0))/1024
                reply['H3C-Input-Average-Rate'] = input_rate
                reply['H3C-Input-Peak-Rate'] = input_rate
                reply['H3C-Output-Average-Rate'] = output_rate
                reply['H3C-Output-Peak-Rate'] = output_rate
        except Exception as err:
            logger.error("rate limit error {}".format(err.message), exc_info=debug)

        return reply



request_parse = RequestParse
rate_limit = RateLimit
kbps_rate_limit = RateLimit


__all__ = ['request_parse', 'rate_limit', 'kbps_rate_limit']