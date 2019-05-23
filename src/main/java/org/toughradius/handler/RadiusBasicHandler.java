package org.toughradius.handler;

import org.apache.mina.core.future.WriteFuture;
import org.toughradius.common.ValidateCache;
import org.toughradius.config.RadiusConfig;
import org.toughradius.entity.Bras;
import org.toughradius.entity.Subscribe;
import org.tinyradius.attribute.RadiusAttribute;
import org.tinyradius.packet.AccessAccept;
import org.tinyradius.packet.AccessRequest;
import org.tinyradius.packet.AccountingRequest;
import org.tinyradius.packet.RadiusPacket;
import org.tinyradius.util.RadiusException;
import org.toughradius.component.RadiusStat;
import org.apache.mina.core.buffer.IoBuffer;
import org.apache.mina.core.service.IoHandlerAdapter;
import org.apache.mina.core.session.IdleStatus;
import org.apache.mina.core.session.IoSession;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor;
import org.toughradius.component.*;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.SocketAddress;
import java.util.*;

public abstract class RadiusBasicHandler extends IoHandlerAdapter {

    protected  final String SESSION_CLIENT_IP_KEY = "SESSION_CLIENT_IP_KEY";
    protected  final String SESSION_TYPE = "SESSION_TYPE";
    protected  final String SESSION_RADSEC_TYPE = "SESSION_RADSEC_TYPE";
    protected  final String SESSION_UDP_TYPE = "SESSION_UDP_TYPE";

    @Autowired
    protected RadiusStat radiusStat;

    @Autowired
    protected RadiusConfig radiusConfig;

    @Autowired
    protected BrasService brasService;

    @Autowired
    protected SubscribeService subscribeService;

    @Autowired
    protected OnlineCache onlineCache;

    @Autowired
    protected SubscribeCache subscribeCache;

    @Autowired
    protected ConfigService configService;

    @Autowired
    protected RadiusParseFilter parseFilter;

    @Autowired
    protected RadiusAuthStat radiusAuthStat;

    @Autowired
    protected RadiusCastStat radiusCastStat;

    @Autowired
    protected ThreadPoolTaskExecutor systaskExecutor;

    @Autowired
    protected RadiusAcceptFilter acceptFilter;

    @Autowired
    protected RadiusAccountingFilter accountingFilter;

    @Autowired
    protected Memarylogger logger;

    private Set<IoSession> sessionSet = new HashSet<>();


    private Map<Long, ValidateCache> authValidateMap = new HashMap<>();

    /**
     * BRAS  Auth 并发限制
     * @param bras
     * @return
     */
    protected ValidateCache getAuthBrasValidate(Bras bras){
        if(authValidateMap.containsKey(bras.getId())){
            ValidateCache vc = authValidateMap.get(bras.getId());
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
            authValidateMap.put(bras.getId(),vc);
            return vc;
        }
    }


    private Map<Long,ValidateCache> acctValidateMap = new HashMap<>();

    /**
     * BRAS Acct 并发限制
     * @param bras
     * @return
     */
    protected ValidateCache getAcctBrasValidate(Bras bras){
        if(acctValidateMap.containsKey(bras.getId())){
            ValidateCache vc = acctValidateMap.get(bras.getId());
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
            acctValidateMap.put(bras.getId(),vc);
            return vc;
        }
    }

    /**
     * 查询设备信息
     * @param client
     * @param packet
     * @return
     * @throws RadiusException
     */
    public Bras getNas(InetSocketAddress client, RadiusPacket packet) throws RadiusException {
        String ip = client.getAddress().getHostAddress();
        RadiusAttribute nasid = packet.getAttribute(32);
        try {
            return brasService.findBras(ip,null,nasid.getAttributeValue());
        } catch (ServiceException e) {
            throw  new RadiusException(e.getMessage());
        }
    }

    /**
     * 查询用户信息
     * @param username
     * @return
     */
    public Subscribe getUser(String username) {
        return subscribeCache.findSubscribe(username);
    }

    /**
     * 验证用户密码
     * @param user
     * @param accessRequest
     * @throws RadiusException
     */
    public void authUser(Subscribe user, AccessRequest accessRequest) throws RadiusException {
        String plaintext = user.getPassword();
        String ignorePwd = configService.getStringValue(ConfigService.RADIUS_MODULE,ConfigService.RADIUS_IGNORE_PASSWORD);

        if(!"enabled".equals(ignorePwd)){
            if (plaintext == null || !accessRequest.verifyPassword(plaintext)){
                radiusAuthStat.update(RadiusAuthStat.PWD_ERR);
                throw new RadiusException("密码错误");
            }
        }
    }

    /**
     * 拷贝代理状态属性
     * @param request
     * @param answer
     */
    protected void copyProxyState(RadiusPacket request, RadiusPacket answer) {
        List proxyStateAttrs = request.getAttributes(33);
        for (Iterator i = proxyStateAttrs.iterator(); i.hasNext();) {
            RadiusAttribute proxyStateAttr = (RadiusAttribute) i.next();
            answer.addAttribute(proxyStateAttr);
        }
    }

    /**
     * 创建记帐响应包
     * @param accountingRequest
     * @return
     * @throws RadiusException
     */
    public RadiusPacket getAccountingResponse(AccountingRequest accountingRequest) throws RadiusException {
        RadiusPacket answer = new RadiusPacket(RadiusPacket.ACCOUNTING_RESPONSE, accountingRequest.getPacketIdentifier());
        copyProxyState(accountingRequest, answer);
        return answer;
    }

    /**
     * 创建认证授权响应
     * @param accessRequest
     * @return
     */
    public AccessAccept getAccessAccept(AccessRequest accessRequest) {
        AccessAccept answer = new AccessAccept(accessRequest.getPacketIdentifier());
        answer.addAttribute("Reply-Message","ok");
        copyProxyState(accessRequest, answer);
        return answer;
    }

    /**
     * 创建认证拒绝响应
     * @param accessRequest
     * @param error
     * @return
     */
    public RadiusPacket getAccessReject(AccessRequest accessRequest, String error) {
        RadiusPacket answer = new RadiusPacket(RadiusPacket.ACCESS_REJECT, accessRequest.getPacketIdentifier());
        if(error==null){
            error = "Unknow Error";
        }
        answer.addAttribute("Reply-Message",error);
        copyProxyState(accessRequest, answer);
        return answer;
    }

    /**
     * 解码原始数据傲文
     * @param data
     * @param sharedSecret
     * @param forceType
     * @return
     * @throws IOException
     * @throws RadiusException
     */
    protected RadiusPacket makeRadiusPacket(byte[] data, String sharedSecret, int forceType) throws IOException, RadiusException {
        ByteArrayInputStream in = new ByteArrayInputStream(data);
        return RadiusPacket.decodeRequestPacket(in, sharedSecret, forceType);
    }

    /**
     * 数据报文解析
     * @param session
     * @param message
     * @return
     * @throws IOException
     * @throws RadiusException
     */
    protected byte[] parseMessage(IoSession session, Object message) throws IOException, RadiusException {
        if (!(message instanceof IoBuffer)) {
            return null;
        }
        IoBuffer buffer = (IoBuffer) message;
        byte[] data = new byte[buffer.limit()];
        buffer.get(data);
        radiusStat.incrReqBytes(data.length);
        return  data;
    }

    /**
     * 发送正常响应
     * @param session
     * @param remoteAddress
     * @param secret
     * @param request
     * @param response
     * @throws IOException
     */
    protected void sendResponse(IoSession session, SocketAddress remoteAddress, String secret, RadiusPacket request, RadiusPacket response) throws IOException {
        ByteArrayOutputStream bos = new ByteArrayOutputStream();
        response.encodeResponsePacket(bos, secret, request);
        byte [] data = bos.toByteArray();
        IoBuffer outbuff = IoBuffer.wrap(data);
        radiusStat.incrRespBytes(data.length);
        if(session.getAttribute(SESSION_TYPE).equals(SESSION_RADSEC_TYPE)){
            session.write(outbuff);
        }else{
            session.write(outbuff,remoteAddress);
            session.closeOnFlush();
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
        if (user.getBindMac()!=null&&user.getBindMac()==1) {
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
        if (user.getBindVlan()!=null&&user.getBindVlan()==1) {
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
        if (user.getBindVlan()!=null&&user.getBindVlan()==1) {
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


    public RadiusPacket accountingRequestReceived(AccountingRequest accountingRequest, Bras nas) throws RadiusException {
        try {
            Subscribe user = getUser(accountingRequest.getUserName());
            accountingFilter.doFilter(accountingRequest,nas,user);
        } catch (RadiusException e) {
            logger.error(accountingRequest.getUserName(),"记账处理错误",e, Memarylogger.RADIUSD);
        }
        return getAccountingResponse(accountingRequest);
    }

    @Override
    public void exceptionCaught(IoSession session, Throwable cause)throws Exception {
        cause.printStackTrace();
        session.closeNow();
        sessionSet.remove(session);
    }

    @Override
    public void sessionClosed(IoSession session) throws Exception {
        sessionSet.remove(session);
    }

    @Override
    public void sessionCreated(IoSession session) throws Exception {
        InetSocketAddress remoteAddress = (InetSocketAddress) session.getRemoteAddress();
        session.setAttribute(SESSION_CLIENT_IP_KEY, remoteAddress);
        session.setAttribute(SESSION_TYPE, SESSION_UDP_TYPE);
        logger.print("UdpSession created " + session.toString());

    }

    @Override
    public void sessionIdle(IoSession session, IdleStatus status) throws Exception {
    }

    @Override
    public void sessionOpened(IoSession session) throws Exception {
    }

    protected void addSession(IoSession session){
        sessionSet.add(session);
    }

    protected Set<IoSession> getSessionSet(){
        return sessionSet;
    }

}
