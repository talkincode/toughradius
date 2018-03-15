#coding=utf-8

from toughradius.modules import VENDOR_ZTE
import logging

logger = logging.getLogger(__name__)
def handle_radius(req, reply, debug=False):
    try:
        if int(req.vendor_id) == VENDOR_ZTE:
            reply['ZTE-Rate-Ctrl-Scr-Up'] = int(reply.resp_attrs.get('input_rate', 0))
            reply['ZTE-Rate-Ctrl-Scr-Down'] = int(reply.resp_attrs.get('output_rate', 0))
    except Exception as err:
        logger.error("response parse error {}".format(err.message),exc_info=debug)

    return reply

