# coding=utf-8

from toughradius.modules import VENDOR_H3C
import logging
logger = logging.getLogger(__name__)

def handle_radius(req, reply, debug=False):
    try:
        if int(req.vendor_id) == VENDOR_H3C:
            if reply.resp_attrs.get('rate_unit', 'bps') == 'kbps':
                input_rate = int(reply.resp_attrs.get('input_rate', 0))/1024
                output_rate = int(reply.resp_attrs.get('output_rate', 0))/1024
            else:
                input_rate = int(reply.resp_attrs.get('input_rate', 0))
                output_rate = int(reply.resp_attrs.get('output_rate', 0))
            reply['H3C-Input-Average-Rate'] = input_rate
            reply['H3C-Input-Peak-Rate'] = input_rate
            reply['H3C-Output-Average-Rate'] = output_rate
            reply['H3C-Output-Peak-Rate'] = output_rate
    except Exception as err:
        logger.error("response parse error {}".format(err.message), exc_info=debug)

    return reply


