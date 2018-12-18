package org.toughradius.handler;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.apache.mina.core.buffer.IoBuffer;
import org.apache.mina.core.session.IoSession;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.tinyradius.packet.AccountingRequest;
import org.tinyradius.packet.RadiusPacket;
import org.tinyradius.util.RadiusException;
import org.toughradius.entity.Nas;
import org.toughradius.entity.User;

import java.net.InetSocketAddress;

@Component
public class RadiusAcctHandler extends RadiusbasicHandler {

    private final static Log logger = LogFactory.getLog(RadiusAcctHandler.class);


    @Autowired
    private RadiusAccountingFilter accountingFilter;


    @Autowired
    private RadiusParseFilter parseFilter;


    public RadiusPacket accountingRequestReceived(AccountingRequest accountingRequest, Nas nas) throws RadiusException {
        schedExecuter.execute(()-> {
            try {
                User user = getUser(accountingRequest.getUserName());
                accountingFilter.doFilter(accountingRequest,nas,user);
            } catch (RadiusException e) {
                logger.error("记账处理错误",e);
            }
        });
        return getAccountingResponse(accountingRequest);
    }

    @Override
    public void messageReceived(IoSession session, Object message)
            throws Exception {

        if (!(message instanceof IoBuffer)) {
            return;
        }
        IoBuffer buffer = (IoBuffer) message;
        byte[] data = new byte[buffer.limit()];
        buffer.get(data);
        radiusStat.incrReqBytes(data.length);

        final InetSocketAddress remoteAddress = (InetSocketAddress) session.getRemoteAddress();
        final InetSocketAddress localAddress = (InetSocketAddress) session.getLocalAddress();
        RadiusPacket preRequest = makeRadiusPacket(data, "1234567890", RadiusPacket.RESERVED);
        radiusStat.incrAcctReq();
        final Nas nas = getNas(remoteAddress, preRequest);

        if (nas == null) {
            radiusStat.incrAcctDrop();
            logger.error(String.format("未授权的接入设备<记账> <%s -> %s>", remoteAddress,localAddress));
            return;
        }

        AccountingRequest request = (AccountingRequest)makeRadiusPacket(data, nas.getSecret(), RadiusPacket.UNDEFINED);
        request.setRemoteAddr(remoteAddress);
//        if (isPacketDuplicate(request, remoteAddress)){
//            logger.error(getClass().toString(),String.format("重复的记账请求 <%s -> %s> : %s", remoteAddress,localAddress,request));
//            return;
//        }

        request = (AccountingRequest)parseFilter.doFilter(request,nas);
        logger.info(String.format("接收到RADIUS 记账请求 <%s -> %s> : %s", remoteAddress,localAddress,request.toSimpleString()));
        if (radiusConfig.isTraceEnabled())
            logger.info(request.toString());

        // handle packet
        final RadiusPacket response = accountingRequestReceived(request, nas);

        // send response
        if (response != null) {
            radiusStat.incrAcctResp();
            logger.info(String.format("发送记账响应至 %s， %s",remoteAddress,response.toLineString()));
            if (radiusConfig.isTraceEnabled())
                logger.info(response.toString());

            sendResponse(session,remoteAddress,nas.getSecret(),request,response);
        }

    }

}
