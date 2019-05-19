package org.toughradius.handler;

import org.toughradius.common.ValidateCache;
import org.toughradius.component.Memarylogger;
import org.toughradius.component.RadiusAuthStat;
import org.toughradius.component.RadiusCastStat;
import org.toughradius.entity.Bras;
import org.toughradius.entity.Subscribe;
import org.tinyradius.packet.AccessAccept;
import org.tinyradius.packet.AccessRequest;
import org.tinyradius.packet.RadiusPacket;
import org.tinyradius.util.RadiusException;
import org.apache.mina.core.session.IoSession;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.net.InetSocketAddress;
import java.util.Date;
import java.util.HashMap;
import java.util.Map;

@Component
public class RadiusAuthHandler extends RadiusBasicHandler {

    @Autowired
    private RadiusAcceptFilter acceptFilter;
    /**
     * BRAS 并发限制
     */
    private Map<Long,ValidateCache> validateMap = new HashMap<>();




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
            radiusAuthStat.update(RadiusAuthStat.NOT_EXIST);
            throw new RadiusException("用户 " + accessRequest.getUserName() + " 不存在");
        }else if("disabled".equals(user.getStatus())){
            radiusAuthStat.update(RadiusAuthStat.STATUS_ERR);
            throw new RadiusException("用户 " + accessRequest.getUserName() + " 已禁用");
        }else if("pause".equals(user.getStatus())){
            radiusAuthStat.update(RadiusAuthStat.STATUS_ERR);
            throw new RadiusException("用户 " + accessRequest.getUserName() + " 已停用");
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

        if (onlineCache.isLimitOver(user.getSubscriber(),user.getActiveNum())) {
            radiusAuthStat.update(RadiusAuthStat.LIMIT_ERR);
            throw new RadiusException("用户在线数超过限制(MAX=" + user.getActiveNum() + ")");
        }

        //判断MAC绑定
        if (user.getBindMac()) {
            if (user.getMacAddr() == null||"".equals(user.getMacAddr())) {
                systaskExecutor.execute(() -> {
                    subscribeService.updateMacAddr(accessRequest.getUserName(), accessRequest.getMacAddr());
                    if(radiusConfig.isTraceEnabled())
                        logger.info(accessRequest.getUserName(), "用户MAC绑定更新：" + accessRequest.getMacAddr());
                });
            } else if (!user.getMacAddr().equals(accessRequest.getMacAddr())) {
                radiusAuthStat.update(RadiusAuthStat.BIND_ERR);
                throw new RadiusException("用户MAC绑定不匹配， 请求MAC =" + accessRequest.getMacAddr() + ", 绑定MAC =" + user.getMacAddr());
            }
        }
        //判断invlan绑定
        if (user.getBindVlan()) {
            if (user.getInVlan() == null || user.getInVlan() == 0) {
                systaskExecutor.execute(() -> {
                    subscribeService.updateInValn(accessRequest.getUserName(), accessRequest.getInVlanId());
                    if(radiusConfig.isTraceEnabled())
                        logger.info(accessRequest.getUserName(), "用户内层VLAN绑定更新：" + accessRequest.getInVlanId());
                });
            } else if (user.getInVlan() != accessRequest.getInVlanId()) {
                radiusAuthStat.update(RadiusAuthStat.BIND_ERR);
                throw new RadiusException("用户内层VLAN绑定不匹配 请求invlan =" + accessRequest.getInVlanId() + ", 绑定invlan =" + user.getInVlan());
            }
        }
        //判断outvlan绑定
        if (user.getBindVlan()) {
            if (user.getOutVlan() == null || user.getOutVlan() == 0) {
                systaskExecutor.execute(() -> {
                    subscribeService.updateOutValn(accessRequest.getUserName(), accessRequest.getOutVlanId());
                    if(radiusConfig.isTraceEnabled())
                        logger.info(accessRequest.getUserName(), "用户外层VLAN绑定更新：" + accessRequest.getOutVlanId());
                });
            } else if (user.getOutVlan() != accessRequest.getOutVlanId()) {
                radiusAuthStat.update(RadiusAuthStat.BIND_ERR);
                throw new RadiusException("用户外层VLAN绑定不匹配 请求outvlan =" + accessRequest.getOutVlanId() + ", 绑定outvlan =" + user.getOutVlan());
            }
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
        byte[] data = parseMessage(session, message);

        radiusStat.incrReqBytes(data.length);

        final InetSocketAddress remoteAddress = (InetSocketAddress) session.getAttribute(SESSION_CLIENT_IP_KEY);
        final InetSocketAddress localAddress = (InetSocketAddress) session.getLocalAddress();
        RadiusPacket preRequest = makeRadiusPacket(data, "1234567890", RadiusPacket.RESERVED);
        if(preRequest.getPacketType()!=RadiusPacket.ACCESS_REQUEST){
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
        AccessRequest request = (AccessRequest)makeRadiusPacket(data, nas.getSecret(), RadiusPacket.ACCESS_REQUEST);
        request.setRemoteAddr(remoteAddress);

        request = (AccessRequest)parseFilter.doFilter(request,nas);

        ValidateCache vc = getBrasValidate(nas);
        String vckey = nas.getId().toString();
        vc.incr(vckey);
        if(vc.isOver(vckey)){
            radiusStat.incrAuthDrop();
            radiusAuthStat.update(RadiusAuthStat.BRAS_LIMIT_ERR);
            logger.error(request.getUsername(), "接入设备认证并发限制超过" + nas.getAuthLimit() + " <" + remoteAddress + " -> " + localAddress + ">", Memarylogger.RADIUSD);
            return;
        }


        logger.info(request.getUsername(), "接收到RADIUS 认证请求 <" + remoteAddress + " -> " + localAddress + "> : " + request.toSimpleString(), Memarylogger.RADIUSD);
        if (radiusConfig.isTraceEnabled())
            logger.print(request.toString());

        // handle packet
        RadiusPacket response = null;
        try{
            response = accessRequestReceived(request, nas);
            radiusStat.incrAuthAccept();
        } catch(Exception e){
            radiusStat.incrAuthReject();
            logger.error(request.getUserName(), "认证处理失败 " + e.getMessage(), Memarylogger.RADIUSD);
            response = getAccessReject(request,e.getMessage());
        }

        // send response
        if (response != null) {
            logger.info(request.getUsername(), "发送认证响应至 " + remoteAddress + "， " + response.toLineString(), Memarylogger.RADIUSD);
            if (radiusConfig.isTraceEnabled())
                logger.print(response.toString());
            sendResponse(session,remoteAddress,nas.getSecret(),request,response);
        }
        int cast = (int) (System.currentTimeMillis()-start);
        radiusCastStat.updateAuth(cast);
    }

}


