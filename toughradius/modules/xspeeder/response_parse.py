# coding=utf-8
from toughradius.modules import VENDOR_XSPEEDER
import logging

logger = logging.getLogger(__name__)

def handle_radius(req, reply, debug=False):
    try:
        if int(req.vendor_id) == VENDOR_XSPEEDER:
            input_rate = int(reply.resp_attrs.get('input_rate', 0)) /1024 /8
            output_rate = int(reply.resp_attrs.get('output_rate', 0)) / 1024/8
            reply['XSpeeder_UpLoad_Speed'] = input_rate
            reply['XSpeeder_DownLoad_Speed'] = output_rate
    except Exception as err:
        logger.error("response parse error {}".format(err.message), exc_info=debug)

    return reply