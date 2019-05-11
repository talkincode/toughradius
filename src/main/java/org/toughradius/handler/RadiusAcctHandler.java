package org.toughradius.handler;

import org.toughradius.common.ValidateCache;
import org.toughradius.component.Memarylogger;
import org.toughradius.entity.Bras;
import org.toughradius.entity.Subscribe;
import org.tinyradius.packet.AccountingRequest;
import org.tinyradius.packet.RadiusPacket;
import org.tinyradius.util.RadiusException;
import org.apache.mina.core.buffer.IoBuffer;
import org.apache.mina.core.session.IoSession;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.net.InetSocketAddress;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.atomic.AtomicInteger;

@Component
public class RadiusAcctHandler extends RadiusbasicHandler {

    @Autowired
    private RadiusAuthHandler authHandler;


    @Autowired
    private RadiusAccountingFilter accountingFilter;


    @Autowired
    private RadiusParseFilter parseFilter;

    private AtomicInteger counter = new AtomicInteger();

    /**
     * BRAS 并发限制
     */
    private Map<Integer,ValidateCache> validateMap = new HashMap<Integer,ValidateCache>();


    private ValidateCache getBrasValidate(Bras bras){
        if(validateMap.containsKey(bras.getId())){
            ValidateCache vc = validateMap.get(bras.getId());
            Integer limit = bras.getAcctLimit();
            if(limit==null){
                limit = 1000;
            }
            if(limit !=vc.getMaxTimes()){
                vc.setMaxTimes(limit);
            }
            return vc;
        }else{
            Integer limit = bras.getAcctLimit();
            if(limit==null){
                limit = 1000;
            }
            ValidateCache vc = new ValidateCache(1000,limit);
            validateMap.put(bras.getId(),vc);
            return vc;
        }
    }


    public RadiusPacket accountingRequestReceived(AccountingRequest accountingRequest, Bras nas) throws RadiusException {
        systaskExecutor.execute(()-> {
            try {
                Subscribe user = getUser(accountingRequest.getUserName());
                accountingFilter.doFilter(accountingRequest,nas,user);
            } catch (RadiusException e) {
                logger.error(accountingRequest.getUserName(),"记账处理错误",e, Memarylogger.RADIUSD);
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

        final InetSocketAddress remoteAddress = (InetSocketAddress) session.getAttribute(SESSION_CLIENT_IP_KEY);
        final InetSocketAddress localAddress = (InetSocketAddress) session.getLocalAddress();
        RadiusPacket preRequest = makeRadiusPacket(data, "1234567890", RadiusPacket.RESERVED);
        if(preRequest.getPacketType()!=RadiusPacket.ACCOUNTING_REQUEST){
            if(preRequest.getPacketType()==RadiusPacket.ACCESS_REQUEST){
                logger.info("ACCT->AUTH-COUNT:"+counter.incrementAndGet(), Memarylogger.RADIUSD);
                buffer.flip();
                authHandler.messageReceived(session, buffer);
                return;
            }else{
                radiusStat.incrAcctDrop();
                logger.error(String.format("错误的 RADIUS 记账消息类型 %s  <%s -> %s>", preRequest.getPacketType(), remoteAddress,localAddress), Memarylogger.RADIUSD);
                return;
            }
        }
        radiusStat.incrAcctReq();
        final Bras nas = getNas(remoteAddress, preRequest);

        if (nas == null) {
            radiusStat.incrAcctDrop();
            logger.error(String.format("未授权的接入设备<记账> <%s -> %s>", remoteAddress,localAddress), Memarylogger.RADIUSD);
            return;
        }

        AccountingRequest request = (AccountingRequest)makeRadiusPacket(data, nas.getSecret(), RadiusPacket.ACCOUNTING_REQUEST);
        request.setRemoteAddr(remoteAddress);

        request = (AccountingRequest)parseFilter.doFilter(request,nas);

        ValidateCache vc = getBrasValidate(nas);
        String vckey = nas.getId().toString();
        vc.incr(vckey);
        if(vc.isOver(vckey)){
            radiusStat.incrAcctDrop();
            logger.error(request.getUsername(),String.format("接入设备记账并发限制超过%s <%s -> %s>", nas.getAcctLimit(), remoteAddress,localAddress), Memarylogger.RADIUSD);
            return;
        }

        logger.info(request.getUserName(), String.format("接收到 RADIUS 记账(%s)请求 <%s -> %s> : %s",
                request.getStatusTypeName(), remoteAddress,localAddress,request.toSimpleString()), Memarylogger.RADIUSD);
        if (radiusConfig.isTraceEnabled())
            logger.print(request.toString());

        // handle packet
        final RadiusPacket response = accountingRequestReceived(request, nas);

        // send response
        if (response != null) {
            radiusStat.incrAcctResp();
            logger.info(request.getUserName(),String.format("发送 RADIUS 记账(%s)响应至 %s， %s",
                    request.getStatusTypeName(), remoteAddress,response.toLineString()), Memarylogger.RADIUSD);
            if (radiusConfig.isTraceEnabled())
                logger.print(response.toString());

            sendResponse(session,remoteAddress,nas.getSecret(),request,response);
        }

    }

}
