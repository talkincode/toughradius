package org.toughradius.handler;

import org.toughradius.common.ValidateCache;
import org.toughradius.component.Syslogger;
import org.toughradius.entity.Bras;
import org.toughradius.entity.Subscribe;
import org.toughradius.entity.SubscribeBill;
import org.tinyradius.packet.AccessAccept;
import org.tinyradius.packet.AccessRequest;
import org.tinyradius.packet.RadiusPacket;
import org.tinyradius.util.RadiusException;
import org.apache.mina.core.buffer.IoBuffer;
import org.apache.mina.core.session.IoSession;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.net.InetSocketAddress;
import java.util.Date;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.atomic.AtomicInteger;

@Component
public class RadiusAuthHandler extends RadiusbasicHandler{

    @Autowired
    private RadiusAcctHandler acctHandler;

    @Autowired
    private RadiusAcceptFilter acceptFilter;

    @Autowired
    private ValidateCache authValidate;

    private AtomicInteger counter = new AtomicInteger();

    /**
     * BRAS 并发限制
     */
    private Map<Integer,ValidateCache> validateMap = new HashMap<Integer,ValidateCache>();


    private ValidateCache getBrasValidate(Bras bras){
        if(validateMap.containsKey(bras.getId())){
            ValidateCache vc = validateMap.get(bras.getId());
            Integer limit = bras.getAuthLimit();
            if(limit==null){
                limit = 1000;
            }
            if(limit !=vc.getMaxTimes()){
                vc.setMaxTimes(limit);
            }
            return vc;
        }else{
            Integer limit = bras.getAuthLimit();
            if(limit==null){
                limit = 1000;
            }
            ValidateCache vc = new ValidateCache(1000,limit);
            validateMap.put(bras.getId(),vc);
            return vc;
        }
    }

    /**
     * 用户认证请求处理
     * @param accessRequest
     * @param nas
     * @return
     * @throws RadiusException
     */
    public RadiusPacket accessRequestReceived(AccessRequest accessRequest, Bras nas) throws RadiusException {
        Subscribe user = getUser(accessRequest.getUserName());

        if(user == null){
            throw new RadiusException(String.format("用户 %s 不存在", accessRequest.getUserName()));
        }else if("disabled".equals(user.getStatus())){
            throw new RadiusException(String.format("用户 %s 已禁用", accessRequest.getUserName()));
        }else if("pause".equals(user.getStatus())){
            throw new RadiusException(String.format("用户 %s 已停用", accessRequest.getUserName()));
        }
        Integer chkpwd = configService.getIsCheckPwd();
        if((chkpwd==null ? 1 : chkpwd)!=0)
            authUser(user, accessRequest);

        long timeout = (user.getExpireTime().getTime() - new Date().getTime())/1000;
        if (timeout <= 0 ) {
            if(radiusConfig.isAllowNegative()){
                timeout = -1;
            }else{
                timeout = 86400;
            }
        }

        //流量判断
        if("flow".equals(user.getBillType())){
            SubscribeBill billdata = subscribeCache.getBillData(accessRequest.getUserName());
            if(billdata.getFlowAmount()==null || billdata.getFlowAmount().longValue() <=0){
                timeout = 86400;
            }
        }


        //判断MAC绑定
        if (user.getBindMac()) {
            if (user.getMacAddr() == null||"".equals(user.getMacAddr())) {
                taskExecutor.execute(() -> {
                    subscribeService.updateMacAddr(accessRequest.getUserName(), accessRequest.getMacAddr());
                    if(radiusConfig.isTraceEnabled())
                        logger.info(accessRequest.getUserName(), String.format("用户MAC绑定更新：%s", accessRequest.getMacAddr()));
                });
            } else if (!user.getMacAddr().equals(accessRequest.getMacAddr())) {
                throw new RadiusException("用户MAC绑定不匹配， 请求MAC =" + accessRequest.getMacAddr() + ", 绑定MAC =" + user.getMacAddr());
            }
        }
        //判断invlan绑定
        if (user.getBindVlan()) {
            if (user.getInVlan() == null || user.getInVlan() == 0) {
                taskExecutor.execute(() -> {
                    subscribeService.updateInValn(accessRequest.getUserName(), accessRequest.getInVlanId());
                    if(radiusConfig.isTraceEnabled())
                        logger.info(accessRequest.getUserName(), String.format("用户内层VLAN绑定更新：%s", accessRequest.getInVlanId()));
                });
            } else if (user.getInVlan() != accessRequest.getInVlanId()) {
                throw new RadiusException("用户内层VLAN绑定不匹配 请求invlan =" + accessRequest.getInVlanId() + ", 绑定invlan =" + user.getInVlan());
            }
        }
        //判断outvlan绑定
        if (user.getBindVlan()) {
            if (user.getOutVlan() == null || user.getOutVlan() == 0) {
                taskExecutor.execute(() -> {
                    subscribeService.updateOutValn(accessRequest.getUserName(), accessRequest.getOutVlanId());
                    if(radiusConfig.isTraceEnabled())
                        logger.info(accessRequest.getUserName(), String.format("用户外层VLAN绑定更新：%s", accessRequest.getOutVlanId()));
                });
            } else if (user.getOutVlan() != accessRequest.getOutVlanId()) {
                throw new RadiusException("用户外层VLAN绑定不匹配 请求outvlan =" + accessRequest.getOutVlanId() + ", 绑定outvlan =" + user.getOutVlan());
            }
        }


        int curCount = onlineCache.getUserOnlineNum(accessRequest.getUserName());
        if (curCount >= user.getActiveNum()) {
            throw new RadiusException(String.format("用户在线数超过限制(MAX=%s)", user.getActiveNum()));
        }

        AccessAccept accept = getAccessAccept(accessRequest);
        accept.setPreSessionTimeout(timeout);
        accept.setPreInterim(radiusConfig.getInterimUpdate());
        accept =   acceptFilter.doFilter(accept,nas,user);
        accessRequest.addMSCHAPV2Response(accept,user,nas);
        return accept;
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
    public void messageReceived(IoSession session, Object message)
            throws Exception {
        long start = System.currentTimeMillis();
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
        if(preRequest.getPacketType()!=RadiusPacket.ACCESS_REQUEST){
            if(preRequest.getPacketType()==RadiusPacket.ACCOUNTING_REQUEST){
                logger.info("AUTH->ACCT-COUNT:"+counter.incrementAndGet(),Syslogger.RADIUSD);
                buffer.flip();
                acctHandler.messageReceived(session,buffer);
                return;
            }else {
                radiusStat.incrAuthDrop();
                logger.error(String.format("错误的 RADIUS 认证消息类型 %s  <%s -> %s>", preRequest.getPacketType(), remoteAddress,localAddress), Syslogger.RADIUSD);
                return;
            }
        }

        radiusStat.incrAuthReq();
        final Bras nas = getNas(remoteAddress, preRequest);

        if (nas == null) {
            radiusStat.incrAuthDrop();
            logger.error(String.format("未授权的接入设备<认证> <%s -> %s>", remoteAddress,localAddress), Syslogger.RADIUSD);
            return;
        }

        // parse packet
        AccessRequest request = (AccessRequest)makeRadiusPacket(data, nas.getSecret(), RadiusPacket.ACCESS_REQUEST);
        request.setRemoteAddr(remoteAddress);

        request = (AccessRequest)parseFilter.doFilter(request,nas);

        ValidateCache vc = getBrasValidate(nas);
        String vckey = nas.getId().toString();
        vc.incr(vckey);
        if(vc.isOver(vckey)){
            radiusStat.incrAuthDrop();
            logger.error(request.getUsername(),String.format("接入设备认证并发限制超过%s <%s -> %s>", nas.getAuthLimit(), remoteAddress,localAddress), Syslogger.RADIUSD);
            return;
        }


        logger.info(request.getUsername(), String.format("接收到RADIUS 认证请求 <%s -> %s> : %s", remoteAddress,localAddress,request.toSimpleString()),Syslogger.RADIUSD);
        if (radiusConfig.isTraceEnabled())
            logger.print(request.toString());

        // handle packet
        RadiusPacket response = null;
        try{
            response = accessRequestReceived(request, nas);
            radiusStat.incrAuthAccept();
        } catch(Exception e){
            radiusStat.incrAuthReject();
            logger.error(request.getUserName(), String.format("认证处理失败 %s", e.getMessage()),Syslogger.RADIUSD);
            response = getAccessReject(request,e.getMessage());
        }

        // send response
        if (response != null) {
            logger.info(request.getUsername(), String.format("发送认证响应至 %s， %s",remoteAddress,response.toLineString()),Syslogger.RADIUSD);
            if (radiusConfig.isTraceEnabled())
                logger.print(response.toString());

            if(response.getPacketType()==RadiusPacket.ACCESS_ACCEPT){
                sendResponse(session,remoteAddress,nas.getSecret(),request,response);
            }else{
                if(radiusConfig.getRejectdelayEnabled() == 1){
                    //检查是否启用拒绝延迟, 为防止ddos攻击，对频繁认证错误的请求延迟响应
                    authValidate.incr(request.getUsername());
                    if(authValidate.isOver(request.getUsername())){
                        radiusStat.incrAuthRejectdelay();
                        sendDelayResponse(radiusConfig.getRejectdelay(),session,remoteAddress,nas.getSecret(),request,response);
                    }else{
                        sendResponse(session,remoteAddress,nas.getSecret(),request,response);
                    }
                }else{
                    sendResponse(session,remoteAddress,nas.getSecret(),request,response);
                }
            }
        }
        if (radiusConfig.isTraceEnabled())
            logger.print(String.format("用户认证处理耗时:%s 毫秒", (System.currentTimeMillis()-start)));
    }

}


