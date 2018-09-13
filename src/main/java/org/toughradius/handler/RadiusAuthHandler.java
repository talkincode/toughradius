package org.toughradius.handler;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.apache.mina.core.buffer.IoBuffer;
import org.apache.mina.core.session.IoSession;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.tinyradius.packet.AccessAccept;
import org.tinyradius.packet.AccessRequest;
import org.tinyradius.packet.RadiusPacket;
import org.tinyradius.util.RadiusException;
import org.toughradius.entity.Nas;
import org.toughradius.entity.User;

import java.math.BigInteger;
import java.net.InetSocketAddress;
import java.util.Date;

@Component
public class RadiusAuthHandler extends RadiusbasicHandler{

    private final static Log logger = LogFactory.getLog(RadiusAuthHandler.class);
    @Autowired
    private RadiusAcceptFilter acceptFilter;

    /**
     * 用户认证请求处理
     * @param accessRequest
     * @param nas
     * @return
     * @throws RadiusException
     */
    public RadiusPacket accessRequestReceived(AccessRequest accessRequest, Nas nas) throws RadiusException {
        User user = getUser(accessRequest.getUserName());

        if(user == null){
            throw new RadiusException(String.format("用户 %s 不存在", accessRequest.getUserName()));
        }else if("disabled".equals(user.getStatus())){
            throw new RadiusException(String.format("用户 %s 已停用", accessRequest.getUserName()));
        }

        if(configService.getIsCheckPwd()!=0)
            authUser(user, accessRequest);

        long timeout = (user.getExpireTime().getTime() - new Date().getTime())/1000;


        //判断MAC绑定
        if (user.getBindMac()==1) {
            if (user.getMacAddr() == null||"".equals(user.getMacAddr())) {
                taskExecutor.execute(() -> {
                    subscribeService.updateMacAddr(accessRequest.getUserName(), accessRequest.getMacAddr());
                    if(radiusConfig.isTraceEnabled())
                        logger.info(String.format("用户MAC绑定更新：%s", accessRequest.getMacAddr()));
                });
            } else if (!user.getMacAddr().equals(accessRequest.getMacAddr())) {
                throw new RadiusException("用户MAC绑定不匹配， 请求MAC =" + accessRequest.getMacAddr() + ", 绑定MAC =" + user.getMacAddr());
            }
        }
        //判断invlan绑定
        if (user.getBindVlan()==1) {
            if (user.getInVlan() == null || user.getInVlan() == 0) {
                taskExecutor.execute(() -> {
                    subscribeService.updateInValn(accessRequest.getUserName(), accessRequest.getInVlanId());
                    if(radiusConfig.isTraceEnabled())
                        logger.info(String.format("用户内层VLAN绑定更新：%s", accessRequest.getInVlanId()));
                });
            } else if (user.getInVlan() != accessRequest.getInVlanId()) {
                throw new RadiusException("用户内层VLAN绑定不匹配 请求invlan =" + accessRequest.getInVlanId() + ", 绑定invlan =" + user.getInVlan());
            }
        }
        //判断outvlan绑定
        if (user.getBindVlan()==1) {
            if (user.getOutVlan() == null || user.getOutVlan() == 0) {
                taskExecutor.execute(() -> {
                    subscribeService.updateOutValn(accessRequest.getUserName(), accessRequest.getOutVlanId());
                    if(radiusConfig.isTraceEnabled())
                        logger.info( String.format("用户外层VLAN绑定更新：%s", accessRequest.getOutVlanId()));
                });
            } else if (user.getOutVlan() != accessRequest.getOutVlanId()) {
                throw new RadiusException("用户外层VLAN绑定不匹配 请求outvlan =" + accessRequest.getOutVlanId() + ", 绑定outvlan =" + user.getOutVlan());
            }
        }

        //流量判断
        if("flow".equals(user.getBillType())){
            BigInteger flowAmount = subscribeCache.getFlowAmount(accessRequest.getUserName());
            if(flowAmount==null || flowAmount.longValue() <=0){
                throw new RadiusException(String.format("用户流量不足, flow_amount = %s", flowAmount.longValue()));
            }
        }

        int curCount = onlineCache.getUserOnlineNum(accessRequest.getUserName());
        if (curCount >= user.getOnlineNum()) {
            throw new RadiusException(String.format("用户在线数超过限制(MAX=%s)", user.getOnlineNum()));
        }

        AccessAccept accept = getAccessAccept(accessRequest);
        accept.setPreSessionTimeout(timeout);
        accept.setPreInterim(radiusConfig.getInterimUpdate());
        return  acceptFilter.doFilter(accept,nas,user);
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

        final InetSocketAddress remoteAddress = (InetSocketAddress) session.getRemoteAddress();
        final InetSocketAddress localAddress = (InetSocketAddress) session.getLocalAddress();
        RadiusPacket preRequest = makeRadiusPacket(data, "1234567890", RadiusPacket.RESERVED);
        radiusStat.incrAuthReq();
        final Nas nas = getNas(remoteAddress, preRequest);

        if (nas == null) {
            radiusStat.incrAuthDrop();
            logger.error(String.format("未授权的接入设备<认证> <%s -> %s>", remoteAddress,localAddress));
            return;
        }

        // parse packet
        AccessRequest request = (AccessRequest)makeRadiusPacket(data, nas.getSecret(), RadiusPacket.UNDEFINED);
        request.setRemoteAddr(remoteAddress);
//        if (isPacketDuplicate(request, remoteAddress)){
//            logger.error(request.getUsername(),String.format("重复的认证请求 <%s -> %s> : %s", remoteAddress,localAddress,request));
//            RadiusPacket reject = getAccessReject(request, "重复的认证消息");
//            sendDelayResponse(radiusConfig.getRejectdelay(), session,remoteAddress,nas.getSecret(),request,reject);
//            return;
//        }

        request = (AccessRequest)parseFilter.doFilter(request,nas);

        logger.info( String.format("接收到RADIUS 认证请求 <%s -> %s> : %s", remoteAddress,localAddress,request.toSimpleString()));
        if (radiusConfig.isTraceEnabled())
            logger.info(request.toString());

        // handle packet
        RadiusPacket response = null;
        try{
            response = accessRequestReceived(request, nas);
            radiusStat.incrAuthAccept();
        } catch(Exception e){
            radiusStat.incrAuthReject();
            logger.error(String.format("认证处理失败 %s", e.getMessage()));
            response = getAccessReject(request,e.getMessage());
        }

        // send response
        if (response != null) {
            logger.info( String.format("发送认证响应至 %s， %s",remoteAddress,response.toLineString()));
            if (radiusConfig.isTraceEnabled())
                logger.info(response.toString());
            sendResponse(session,remoteAddress,nas.getSecret(),request,response);
        }
        if (radiusConfig.isTraceEnabled())
            logger.info(String.format("用户认证处理耗时:%s 毫秒", (System.currentTimeMillis()-start)));
    }

}


