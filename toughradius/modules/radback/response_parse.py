#coding=utf-8
from toughradius.modules import VENDOR_RADBACK
import logging

logger = logging.getLogger(__name__)

def handle_radius(req, reply, debug=False):
    try:
        if int(req.vendor_id) == VENDOR_RADBACK:
            rate_code = reply.resp_attrs.get('rate_code')
            if rate_code:
                reply['Sub-Profile-Name'] = str(rate_code)
    except Exception as err:
        logger.error("response parse error {}".format(err.message), exc_info=debug)

    return reply

