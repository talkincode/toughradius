#coding=utf-8

from toughradius.modules import  VENDOR_IKUAI
import logging

logger = logging.getLogger(__name__)


def handle_radius(req, reply, debug=False):
    try:
        if int(req.vendor_id) == VENDOR_IKUAI:
            input_rate = int(reply.resp_attrs.get('input_rate', 0)) /1024/8
            output_rate = int(reply.resp_attrs.get('output_rate', 0))/1024/8
            reply['RP-Upstream-Speed-Limit'] = input_rate
            reply['RP-Downstream-Speed-Limit'] = output_rate
    except Exception as err:
        logger.error("rate limit error {}".format(err.message), exc_info=debug)

    return reply

