package org.toughradius.handler;

import org.toughradius.common.ValidateCache;
import org.toughradius.component.Memarylogger;
import org.toughradius.component.RadiusCastStat;
import org.toughradius.entity.Bras;
import org.tinyradius.packet.AccountingRequest;
import org.tinyradius.packet.RadiusPacket;
import org.apache.mina.core.session.IoSession;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.net.InetSocketAddress;
import java.util.HashMap;
import java.util.Map;

@Component
public class RadiusAcctHandler extends RadiusBasicHandler {



    @Override
    public void messageReceived(IoSession session, Object message)
            throws Exception {
        long start = System.currentTimeMillis();
        final InetSocketAddress remoteAddress = (InetSocketAddress) session.getAttribute(SESSION_CLIENT_IP_KEY);
        final InetSocketAddress localAddress = (InetSocketAddress) session.getLocalAddress();
        byte[] data = parseMessage(session, message);
        radiusStat.incrReqBytes(data.length);
        RadiusPacket preRequest = makeRadiusPacket(data, "1234567890", RadiusPacket.RESERVED);
        if(preRequest.getPacketType()!=RadiusPacket.ACCOUNTING_REQUEST){
            radiusStat.incrAcctDrop();
            logger.error("错误的 RADIUS 记账消息类型 " + preRequest.getPacketType() + "  <" + remoteAddress + " -> " + localAddress + ">", Memarylogger.RADIUSD);
            return;
        }
        radiusStat.incrAcctReq();
        final Bras nas = getNas(remoteAddress, preRequest);

        if (nas == null) {
            radiusStat.incrAcctDrop();
            logger.error("未授权的接入设备<记账> <" + remoteAddress + " -> " + localAddress + ">", Memarylogger.RADIUSD);
            return;
        }

        AccountingRequest request = (AccountingRequest)makeRadiusPacket(data, nas.getSecret(), RadiusPacket.ACCOUNTING_REQUEST);
        request.setRemoteAddr(remoteAddress);
        request = (AccountingRequest)parseFilter.doFilter(request,nas);

        ValidateCache vc = getAcctBrasValidate(nas);
        String vckey = nas.getId().toString();
        vc.incr(vckey);
        if(vc.isOver(vckey)){
            radiusStat.incrAcctDrop();
            logger.error(request.getUsername(), "接入设备记账并发限制超过" + nas.getAcctLimit() + " <" + remoteAddress + " -> " + localAddress + ">", Memarylogger.RADIUSD);
            return;
        }

        logger.info(request.getUserName(), "接收到 RADIUS 记账(" + request.getStatusTypeName() + ")请求 <" + remoteAddress + " -> " + localAddress + "> : " + request.toLineString(), Memarylogger.RADIUSD);
        if (radiusConfig.isTraceEnabled())
            logger.print(request.toString());

        // handle packet
        final RadiusPacket response = accountingRequestReceived(request, nas);

        // send response
        if (response != null) {
            radiusStat.incrAcctResp();
            logger.info(request.getUserName(), "发送 RADIUS 记账(" + request.getStatusTypeName() + ")响应至 " + remoteAddress + "， " + response.toLineString(), Memarylogger.RADIUSD);
            if (radiusConfig.isTraceEnabled())
                logger.print(response.toString());

            sendResponse(session,remoteAddress,nas.getSecret(),request,response);

            int cast = (int) (System.currentTimeMillis()-start);
            if(request.getAcctStatusType()==AccountingRequest.ACCT_STATUS_TYPE_START){
                radiusCastStat.updateAcctStart(cast);
            }else if(request.getAcctStatusType()==AccountingRequest.ACCT_STATUS_TYPE_INTERIM_UPDATE){
                radiusCastStat.updateAcctUpdate(cast);
            }else if(request.getAcctStatusType()==AccountingRequest.ACCT_STATUS_TYPE_STOP){
                radiusCastStat.updateAcctStop(cast);
            }
        }

    }

}
