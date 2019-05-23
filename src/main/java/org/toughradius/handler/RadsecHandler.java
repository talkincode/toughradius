package org.toughradius.handler;

import org.apache.mina.core.buffer.IoBuffer;
import org.apache.mina.core.session.IoSession;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import org.tinyradius.packet.AccessRequest;
import org.tinyradius.packet.AccountingRequest;
import org.tinyradius.packet.CoaRequest;
import org.tinyradius.packet.RadiusPacket;
import org.toughradius.common.ValidateCache;
import org.toughradius.component.Memarylogger;
import org.toughradius.component.RadiusAuthStat;
import org.toughradius.entity.Bras;
import org.toughradius.entity.RadiusTicket;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.net.InetSocketAddress;
import java.util.NoSuchElementException;
import java.util.concurrent.ConcurrentLinkedDeque;


@Component
public class RadsecHandler  extends RadiusBasicHandler {

    @Override
    public void sessionCreated(IoSession session) throws Exception {
        InetSocketAddress remoteAddress = (InetSocketAddress) session.getRemoteAddress();
        session.setAttribute(SESSION_CLIENT_IP_KEY, remoteAddress);
        session.setAttribute(SESSION_TYPE, SESSION_RADSEC_TYPE);
        this.addSession(session);
        logger.print("RadsecSession created " + session.toString());
    }


    @Scheduled(fixedDelay = 100)
    public void sendRadsecCoa()  {
        CoaRequest req;
        while(true){
            try {
                req = onlineCache.peekCoaRequest();
            }catch (NoSuchElementException ne){
                return;
            }
            ByteArrayOutputStream bos = new ByteArrayOutputStream();
            try {
                req.encodeRequestPacket(bos, "radsec");
            } catch (IOException e) {
                e.printStackTrace();
                return;
            }
            byte[] data = bos.toByteArray();
            for (IoSession session : getSessionSet()){
                try{
                    session.write(IoBuffer.wrap(data));
                    return;
                }catch (Exception ignore){

                }
            }
        }
    }

    @Override
    public void messageReceived(IoSession session, Object message)  throws Exception {
        long start = System.currentTimeMillis();
        byte[] data = parseMessage(session, message);
        radiusStat.incrReqBytes(data.length);
        final InetSocketAddress remoteAddress = (InetSocketAddress) session.getAttribute(SESSION_CLIENT_IP_KEY);
        final InetSocketAddress localAddress = (InetSocketAddress) session.getLocalAddress();
        RadiusPacket request = makeRadiusPacket(data, "radsec", RadiusPacket.UNDEFINED);
        if(request.getPacketType()==RadiusPacket.ACCESS_REQUEST){
            AccessRequest accessRequest = (AccessRequest) request;
            this.handlerAccessRequest(session,accessRequest);
            int cast = (int) (System.currentTimeMillis()-start);
            radiusCastStat.updateAuth(cast);
        }else if(request.getPacketType()==RadiusPacket.ACCOUNTING_REQUEST){
            AccountingRequest accountingRequest = (AccountingRequest) request;
            this.handlerAccountingRequest(session,accountingRequest);
            int cast = (int) (System.currentTimeMillis()-start);
            if(accountingRequest.getAcctStatusType()==AccountingRequest.ACCT_STATUS_TYPE_START){
                radiusCastStat.updateAcctStart(cast);
            }else if(accountingRequest.getAcctStatusType()==AccountingRequest.ACCT_STATUS_TYPE_INTERIM_UPDATE){
                radiusCastStat.updateAcctUpdate(cast);
            }else if(accountingRequest.getAcctStatusType()==AccountingRequest.ACCT_STATUS_TYPE_STOP){
                radiusCastStat.updateAcctStop(cast);
            }
        }else if(request.getPacketType()==RadiusPacket.DISCONNECT_ACK){
            logger.info("接收到 RADSEC DISCONNECT_ACK <" + remoteAddress + " -> " + localAddress + ">" + request.toLineString(),Memarylogger.RADIUSD_COA);
        }else if(request.getPacketType()==RadiusPacket.DISCONNECT_NAK){
            logger.info("接收到 RADSEC DISCONNECT_NAK <" + remoteAddress + " -> " + localAddress + ">" +request.toLineString(),Memarylogger.RADIUSD_COA);
        }else if(request.getPacketType()==RadiusPacket.COA_ACK){
            logger.info("接收到 RADSEC COA_ACK <" + remoteAddress + " -> " + localAddress + ">" +request.toLineString(),Memarylogger.RADIUSD_COA);
        }else if(request.getPacketType()==RadiusPacket.COA_NAK){
            logger.info("接收到 RADSEC COA_NAK <" + remoteAddress + " -> " + localAddress + ">" +request.toLineString(),Memarylogger.RADIUSD_COA);
        }else if(request.getPacketType()==RadiusPacket.ACCESS_CHALLENGE){
            logger.info("接收到 RADSEC ACCESS_CHALLENGE <" + remoteAddress + " -> " + localAddress + ">" +request.toLineString(),Memarylogger.RADIUSD_COA);
        }

    }

    private void handlerAccessRequest(IoSession session, AccessRequest request)  throws Exception {
        final InetSocketAddress remoteAddress = (InetSocketAddress) session.getAttribute(SESSION_CLIENT_IP_KEY);
        final InetSocketAddress localAddress = (InetSocketAddress) session.getLocalAddress();
        radiusStat.incrAuthReq();
        final Bras nas = getNas(remoteAddress, request);

        if (nas == null) {
            radiusStat.incrAuthDrop();
            radiusAuthStat.update(RadiusAuthStat.DROP);
            logger.error("未授权的 RADSEC 接入设备<认证> <" + remoteAddress + " -> " + localAddress + ">", Memarylogger.RADIUSD);
            return;
        }

        // parse packet
        request.setRemoteAddr(remoteAddress);

        request = (AccessRequest) parseFilter.doFilter(request,nas);

        ValidateCache vc = getAuthBrasValidate(nas);
        String vckey = nas.getId().toString();
        vc.incr(vckey);
        if(vc.isOver(vckey)){
            radiusStat.incrAuthDrop();
            radiusAuthStat.update(RadiusAuthStat.BRAS_LIMIT_ERR);
            logger.error(request.getUsername(), "RADSEC 接入设备认证并发限制超过" + nas.getAuthLimit() + " <" + remoteAddress + " -> " + localAddress + ">", Memarylogger.RADIUSD);
            return;
        }


        logger.info(request.getUsername(), "接收到RADIUS RADSEC 认证请求 <" + remoteAddress + " -> " + localAddress + "> : " + request.toSimpleString(), Memarylogger.RADIUSD);
        if (radiusConfig.isTraceEnabled())
            logger.print(request.toString());

        // handle packet
        RadiusPacket response = null;
        try{
            response = accessRequestReceived(request, nas);
            radiusStat.incrAuthAccept();
            radiusAuthStat.update(RadiusAuthStat.ACCEPT);
        } catch(Exception e){
            radiusStat.incrAuthReject();
            logger.error(request.getUserName(), "RADSEC 认证处理失败 " + e.getMessage(), Memarylogger.RADIUSD);
            response = getAccessReject(request, "RADSEC 认证处理失败");
        }

        // send response
        if (response != null) {
            logger.info(request.getUsername(), "发送 RADSEC 认证响应至 " + remoteAddress + "， " + response.toLineString(), Memarylogger.RADIUSD);
            if (radiusConfig.isTraceEnabled())
                logger.print(response.toString());
            sendResponse(session,remoteAddress,nas.getSecret(),request,response);
        }
    }

    private void handlerAccountingRequest(IoSession session, AccountingRequest request)  throws Exception {
        final InetSocketAddress remoteAddress = (InetSocketAddress) session.getAttribute(SESSION_CLIENT_IP_KEY);
        final InetSocketAddress localAddress = (InetSocketAddress) session.getLocalAddress();
        radiusStat.incrAcctReq();
        final Bras nas = getNas(remoteAddress, request);

        if (nas == null) {
            radiusStat.incrAcctDrop();
            logger.error("未授权的 RADSEC 接入设备<记账> <" + remoteAddress + " -> " + localAddress + ">", Memarylogger.RADIUSD);
            return;
        }

        request.setRadsec(true);
        request.setRemoteAddr(remoteAddress);
        request = (AccountingRequest)parseFilter.doFilter(request,nas);

        ValidateCache vc = getAcctBrasValidate(nas);
        String vckey = nas.getId().toString();
        vc.incr(vckey);
        if(vc.isOver(vckey)){
            radiusStat.incrAcctDrop();
            logger.error(request.getUsername(), "RADSEC 接入设备记账并发限制超过" + nas.getAcctLimit() + " <" + remoteAddress + " -> " + localAddress + ">", Memarylogger.RADIUSD);
            return;
        }

        logger.info(request.getUserName(), "接收到 RADIUS RADSEC 记账(" + request.getStatusTypeName() + ")请求 <" + remoteAddress + " -> " + localAddress + "> : " + request.toLineString(), Memarylogger.RADIUSD);
        if (radiusConfig.isTraceEnabled())
            logger.print(request.toString());

        // handle packet
        final RadiusPacket response = accountingRequestReceived(request, nas);
        // send response
        if (response != null) {
            radiusStat.incrAcctResp();
            logger.info(request.getUserName(), "发送 RADIU RADSEC 记账(" + request.getStatusTypeName() + ")响应至 " + remoteAddress + "， " + response.toLineString(), Memarylogger.RADIUSD);
            if (radiusConfig.isTraceEnabled())
                logger.print(response.toString());
            sendResponse(session,remoteAddress,nas.getSecret(),request,response);
        }
    }


}
