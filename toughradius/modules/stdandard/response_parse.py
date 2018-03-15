# coding=utf-8
import logging
from toughradius.common import tools
logger = logging.getLogger(__name__)


def handle_radius(req, reply, debug=False):
    for ak, av in reply.resp_attrs['radius_attrs'].iteritems():
        try:
            if reply.dict.has_key(ak):
                typ = reply.dict[ak].type
                if typ == 'integer' or typ == 'date':
                    reply.AddAttribute(tools.safestr(ak), int(av))
                else:
                    reply.AddAttribute(tools.safestr(ak), str(av))
            else:
                logger.error("unknow radius attr {0}={1}".format(ak,av))
        except Exception as err:
            logger.error("set radius attr error {}".format(err.message), exc_info=debug)
    return reply