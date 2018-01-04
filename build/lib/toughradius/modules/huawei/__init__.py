# coding=utf-8

from toughradius.modules import parse_std_vlan
from toughradius.modules import parse_std_mac
from toughradius.modules import VENDOR_HUAWEI
import logging

logger = logging.getLogger(__name__)


class RequestParse(object):

    parse_vlan = parse_std_vlan
    parse_mac = parse_std_mac

    @staticmethod
    def handle_radius(req, debug=False):
        try:
            if int(req.vendor_id) == VENDOR_HUAWEI:
                req = RequestParse.parse_mac(req)
                req = RequestParse.parse_vlan(req)
        except Exception as err:
            logger.error("request parse error {}".format(err.message), exc_info=debug)

        return req

class RateLimit(object):

    @staticmethod
    def handle_radius(req, reply, debug=False):
        try:
            if int(req.vendor_id) == VENDOR_HUAWEI:
                input_rate = int(reply.resp_attrs.get('input_rate', 0))
                output_rate = int(reply.resp_attrs.get('output_rate', 0))
                reply['Huawei-Input-Average-Rate'] = input_rate
                reply['Huawei-Input-Peak-Rate'] = input_rate
                reply['Huawei-Output-Average-Rate'] = output_rate
                reply['Huawei-Output-Peak-Rate'] = output_rate
        except Exception as err:
            logger.error("rate limit error {}".format(err.message), exc_info=debug)

        return reply

class KbpsRateLimit(object):

    @staticmethod
    def handle_radius(req, reply, debug=False):
        try:
            if int(req.vendor_id) == VENDOR_HUAWEI:
                input_rate = int(reply.resp_attrs.get('input_rate', 0))/1024
                output_rate = int(reply.resp_attrs.get('output_rate', 0))/1024
                reply['Huawei-Input-Average-Rate'] = input_rate
                reply['Huawei-Input-Peak-Rate'] = input_rate
                reply['Huawei-Output-Average-Rate'] = output_rate
                reply['Huawei-Output-Peak-Rate'] = output_rate
        except Exception as err:
            logger.error("rate limit error {}".format(err.message), exc_info=debug)

        return reply



request_parse = RequestParse
rate_limit = RateLimit
kbps_rate_limit = RateLimit