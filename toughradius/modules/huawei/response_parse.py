# coding=utf-8

from toughradius.modules import parse_std_vlan
from toughradius.modules import parse_std_mac
from toughradius.modules import VENDOR_HUAWEI
import logging

logger = logging.getLogger(__name__)


def handle_radius(req, reply, debug=False):
    try:
        if int(req.vendor_id) == VENDOR_HUAWEI:
            if reply.resp_attrs.get('rate_unit', 'bps') == 'kbps':
                input_rate = int(reply.resp_attrs.get('input_rate', 0))/1024
                output_rate = int(reply.resp_attrs.get('output_rate', 0))/1024
            else:
                input_rate = int(reply.resp_attrs.get('input_rate', 0))
                output_rate = int(reply.resp_attrs.get('output_rate', 0))
            reply['Huawei-Input-Average-Rate'] = input_rate
            reply['Huawei-Input-Peak-Rate'] = input_rate
            reply['Huawei-Output-Average-Rate'] = output_rate
            reply['Huawei-Output-Peak-Rate'] = output_rate
    except Exception as err:
        logger.error("response parse error {}".format(err.message), exc_info=debug)

    return reply
