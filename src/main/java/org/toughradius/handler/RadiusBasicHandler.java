package org.toughradius.handler;

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
import java.util.Iterator;
import java.util.List;

public abstract class RadiusBasicHandler extends IoHandlerAdapter {

    protected  final String SESSION_CLIENT_IP_KEY = "SESSION_CLIENT_IP_KEY";

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
    protected ThreadPoolTaskExecutor systaskExecutor;

    @Autowired
    protected Memarylogger logger;

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
        session.write(outbuff,remoteAddress);
        session.closeOnFlush();
    }

    @Override
    public void exceptionCaught(IoSession session, Throwable cause)throws Exception {
        cause.printStackTrace();
        session.closeNow();
    }

    @Override
    public void sessionClosed(IoSession session) throws Exception {
    }

    @Override
    public void sessionCreated(IoSession session) throws Exception {
        InetSocketAddress remoteAddress = (InetSocketAddress) session.getRemoteAddress();
        session.setAttribute(SESSION_CLIENT_IP_KEY, remoteAddress);
    }

    @Override
    public void sessionIdle(IoSession session, IdleStatus status) throws Exception {
    }

    @Override
    public void sessionOpened(IoSession session) throws Exception {
    }
}
