# coding=utf-8
from toughradius.modules import VENDOR_MIKROTIC
import logging

logger = logging.getLogger(__name__)

def handle_radius(req, reply, debug=False):
    try:
        if int(req.vendor_id) == VENDOR_MIKROTIC:
            input_rate = int(reply.resp_attrs.get('input_rate', 0)) /1024
            output_rate = int(reply.resp_attrs.get('output_rate', 0)) / 1024
            reply['Mikrotik-Rate-Limit'] = '%sk/%sk' % (input_rate, output_rate)
    except Exception as err:
        logger.error("rate limit error {}".format(err.message), exc_info=debug)

    return reply