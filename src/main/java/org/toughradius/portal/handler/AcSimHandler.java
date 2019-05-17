package org.toughradius.portal.handler;

import org.apache.mina.core.buffer.IoBuffer;
import org.apache.mina.core.service.IoHandlerAdapter;
import org.apache.mina.core.session.IoSession;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.toughradius.component.BrasService;
import org.toughradius.component.Memarylogger;
import org.toughradius.component.ServiceException;
import org.toughradius.config.PortalConfig;
import org.toughradius.entity.Bras;
import org.toughradius.portal.packet.PortalAttribute;
import org.toughradius.portal.packet.PortalPacket;
import org.toughradius.portal.utils.PortalUtils;

import java.net.InetSocketAddress;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;


@Component
public class AcSimHandler extends IoHandlerAdapter {

    private final String PORTAL_SESSION_CLIENT_IP_KEY = "PORTAL_SESSION_CLIENT_IP_KEY";

    @Autowired
    private Memarylogger logger;

    @Autowired
    protected BrasService brasService;

    @Autowired
    protected PortalConfig portalConfig;

    private Map<Short,byte[]> challengeMap = new ConcurrentHashMap<Short,byte[]>();

    @Override
    public void sessionCreated(IoSession session) throws Exception {
        InetSocketAddress remoteAddress = (InetSocketAddress) session.getRemoteAddress();
        session.setAttribute(PORTAL_SESSION_CLIENT_IP_KEY, remoteAddress);
    }

    /**
     * 异常处理
     * @param session
     * @param cause
     * @throws Exception
     */
    @Override
    public void exceptionCaught(IoSession session, Throwable cause)throws Exception {
        cause.printStackTrace();
        session.closeNow();
    }

    @Override
    public void messageReceived(IoSession session, Object message)throws Exception {
        if (!(message instanceof IoBuffer)) {
            return;
        }
        IoBuffer buffer = (IoBuffer) message;
        byte[] data = new byte[buffer.limit()];
        buffer.get(data);
        PortalPacket request = null;

        final InetSocketAddress remoteAddress = (InetSocketAddress) session.getAttribute(PORTAL_SESSION_CLIENT_IP_KEY);

        Bras nas = null;
        try {
            nas = brasService.findBras("127.0.0.1",null,"inner-tester");
        } catch (ServiceException e) {
            logger.error("【模拟】AC 设备 {nasid=inner-tester，ip=27.0.0.1} 不存在",e,Memarylogger.PORTAL);
            return;
        }


        try{
            request = new PortalPacket(data);
            logger.info("【模拟】接收到 Portal 请求 " + request.toString(),Memarylogger.PORTAL);
        } catch(Exception e){
            logger.error("【模拟】Portal 消息解码失败",e,Memarylogger.PORTAL);
            return;
        }

        PortalPacket resp = null;

        try{
            switch (request.getType()){
                case PortalPacket.REQ_CHALLENGE:{
                    resp = chellengeHandler(request,nas);
                };
                case PortalPacket.REQ_AUTH:{
                    resp = authHandler(request,nas);
                };
                case PortalPacket.REQ_LOGOUT:{
                    resp = logoutHandler(request,nas);
                };
            }
        }catch(Exception e){
            logger.error("【模拟】Portal 消息处理失败",e,Memarylogger.PORTAL);
            return;
        }

        if(resp!=null){
            logger.info(String.format("【模拟】发送 Portal 响应 %s", resp.toString()),Memarylogger.PORTAL);
            IoBuffer outbuff = resp.encodePacket();
            session.write(outbuff,remoteAddress);
            session.closeOnFlush();
        }

    }


    public PortalPacket chellengeHandler(PortalPacket request, Bras nas){
        PortalPacket resp = PortalPacket.createMessage(PortalPacket.getVerbyName(nas.getPortalVendor()),
                PortalPacket.ACK_CHALLENGE,
                request.getUserIp(),
                request.getSerialNo(),
                PortalPacket.getNextReqId(),
                nas.getSecret(),
                portalConfig.getPapchap());
        resp.addAttribute(new PortalAttribute(PortalPacket.ATTRIBUTE_CHALLENGE_TYPE,new byte[16]));
        resp.updateResponseAuthenticator(nas.getSecret());
        return resp;
    }

    public PortalPacket authHandler(PortalPacket request, Bras nas){
        PortalPacket resp = PortalPacket.createMessage(PortalPacket.getVerbyName(nas.getPortalVendor()),
                PortalPacket.ACK_AUTH,
                request.getUserIp(),
                request.getSerialNo(),
                request.getReqId(),
                nas.getSecret(),
                portalConfig.getPapchap());
        resp.addAttribute(new PortalAttribute(PortalPacket.ATTRIBUTE_TEXT_INFO_TYPE, PortalUtils.encodeString("success")));
        resp.updateResponseAuthenticator(nas.getSecret());
        return resp;
    }

    public PortalPacket logoutHandler(PortalPacket request,Bras nas){
        PortalPacket resp = PortalPacket.createMessage(PortalPacket.getVerbyName(nas.getPortalVendor()),
                PortalPacket.ACK_LOGOUT,
                request.getUserIp(),
                request.getSerialNo(),
                request.getReqId(),
                nas.getSecret(),
                portalConfig.getPapchap());
        resp.addAttribute(new PortalAttribute(PortalPacket.ATTRIBUTE_TEXT_INFO_TYPE,PortalUtils.encodeString("success")));
        resp.updateResponseAuthenticator(nas.getSecret());
        return resp;
    }


}
