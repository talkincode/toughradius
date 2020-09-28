package org.toughradius.handler;

import org.tinyradius.util.RadiusException;
import org.toughradius.common.ValidateCache;
import org.toughradius.component.ConfigService;
import org.toughradius.component.Memarylogger;
import org.toughradius.component.RadiusAuthStat;
import org.toughradius.entity.Bras;
import org.tinyradius.packet.AccessRequest;
import org.tinyradius.packet.RadiusPacket;
import org.apache.mina.core.session.IoSession;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.Executors;
import java.util.concurrent.ThreadPoolExecutor;

@Component
public class RadiusAuthHandler extends RadiusBasicHandler {


    /**
     * 异常处理
     *
     * @param session
     * @param cause
     * @throws Exception
     */
    @Override
    public void exceptionCaught(IoSession session, Throwable cause) throws Exception {
        cause.printStackTrace();
        session.closeNow();
    }

    @Override
    public void messageReceived(IoSession session, Object message)
            throws Exception {
        long start = System.currentTimeMillis();
        byte[] data = parseMessage(session, message);

        radiusStat.incrReqBytes(data.length);

        final InetSocketAddress remoteAddress = (InetSocketAddress) session.getAttribute(SESSION_CLIENT_IP_KEY);
        final InetSocketAddress localAddress = (InetSocketAddress) session.getLocalAddress();

        RadiusPacket preRequest = makeRadiusPacket(data, "1234567890", RadiusPacket.RESERVED);
        if (preRequest.getPacketType() != RadiusPacket.ACCESS_REQUEST) {
            radiusStat.incrAuthDrop();
            radiusAuthStat.update(RadiusAuthStat.DROP);
            logger.error("错误的 RADIUS 认证消息类型 " + preRequest.getPacketType() + "  <" + remoteAddress + " -> " + localAddress + ">", Memarylogger.RADIUSD);
            return;
        }

        radiusStat.incrAuthReq();
        final Bras nas = getNas(remoteAddress, preRequest);

        if (nas == null) {
            radiusStat.incrAuthDrop();
            radiusAuthStat.update(RadiusAuthStat.DROP);
            logger.error("未授权的接入设备<认证> <" + remoteAddress + " -> " + localAddress + ">", Memarylogger.RADIUSD);
            return;
        }

        // parse packet
        AccessRequest request = null;
        try {
            request = (AccessRequest) makeRadiusPacket(data, nas.getSecret(), RadiusPacket.ACCESS_REQUEST);
        }catch(RadiusException ex){
            logger.error( "RADIUS 请求解析失败",ex, Memarylogger.RADIUSD);
            return;
        }
        request.setRemoteAddr(remoteAddress);

        request = (AccessRequest) parseFilter.doFilter(request, nas);

        logger.info(request.getUsername(), "接收到RADIUS 认证请求 <" + remoteAddress + " -> " + localAddress + "> : " + request.toSimpleString(), Memarylogger.RADIUSD);
        if (radiusConfig.isTraceEnabled())
            logger.print(request.toString());

        ValidateCache vc = getAuthBrasValidate(nas);
        String vckey = nas.getId().toString();
        vc.incr(vckey);
        if (vc.isOver(vckey)) {
            radiusStat.incrAuthDrop();
            radiusAuthStat.update(RadiusAuthStat.BRAS_LIMIT_ERR);
            logger.error(request.getUsername(), "接入设备认证并发限制超过" + nas.getAuthLimit() + " <" + remoteAddress + " -> " + localAddress + ">", Memarylogger.RADIUSD);
            sendResponse(session, remoteAddress, nas, request, getAccessReject(request, "接入设备认证并发限制超过"));
            return;
        }

        // handle packet
        RadiusPacket response = null;
        try {
            response = accessRequestReceived(request, nas);
            radiusStat.incrAuthAccept();
            radiusAuthStat.update(RadiusAuthStat.ACCEPT);
        } catch (Exception e) {
            radiusStat.incrAuthReject();
            logger.error(request.getUserName(), "认证处理失败 " + e.getMessage(), Memarylogger.RADIUSD);
            response = getAccessReject(request, "认证处理失败");
        }

        sendResponse(session, remoteAddress, nas, request, response);
        int cast = (int) (System.currentTimeMillis() - start);
        radiusCastStat.updateAuth(cast);


    }

    private void sendResponse(IoSession session, InetSocketAddress remoteAddress, Bras nas, AccessRequest request, RadiusPacket response) throws IOException {
        // send response
        if (response != null) {
            logger.info(request.getUsername(), "发送认证响应至 " + remoteAddress + "， " + response.toLineString(), Memarylogger.RADIUSD);
            if (radiusConfig.isTraceEnabled())
                logger.print(response.toString());
            sendResponse(session, remoteAddress, nas.getSecret(), request, response);
        }
    }

}


