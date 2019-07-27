package org.toughradius.component;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.tinyradius.util.RadiusException;
import org.toughradius.common.CoderUtil;
import org.toughradius.common.DateTimeUtil;
import org.toughradius.config.RadiusConfig;
import org.toughradius.entity.Bras;
import org.toughradius.entity.RadiusOnline;
import org.toughradius.entity.RadiusTicket;
import org.toughradius.entity.Subscribe;
import org.toughradius.form.FreeradiusAcctRequest;

import java.util.Date;

@Component
public class FreeradiusService {


    @Autowired
    protected RadiusStat radiusStat;
    @Autowired
    public Memarylogger logger;
    @Autowired
    private OnlineCache onlineCache;
    @Autowired
    private TicketCache ticketCache;
    @Autowired
    private RadiusConfig radiusConfig;


    /**
     * RADIUS 记账处理
     * @param request
     * @param nas
     * @param user
     * @throws RadiusException
     */
    public void doFilter(FreeradiusAcctRequest request, Bras nas, Subscribe user) throws RadiusException {
        String type = request.getAcctStatusType();
        if ("Start".equals(type) ) {
            radiusStat.incrAcctStart();
            doStart(request,nas,user);
        }else if ("Alive".equals(type) || "Interim-Update".equals(type) ) {
            radiusStat.incrAcctUpdate();
            doUpdate(request,nas,user);
        }else if ("Stop".equals(type)) {
            radiusStat.incrAcctStop();
            doStop(request,nas,user);
        }else{
            logger.error(request.getUsername(),String.format("忽略的记账请求类型 %s", type), Memarylogger.RADIUSD);
        }
    }

    /**
     * 新增在线用户
     * @param request
     * @param nas
     * @param user
     * @throws RadiusException
     */
    private void addOnline(FreeradiusAcctRequest request, Bras nas, Subscribe user) throws RadiusException {
        if(user==null){
            String errMessage = "用户不存在或状态未启用（记账请求） " + request.getAcctStatusType();
            logger.error(request.getUsername(), errMessage, Memarylogger.RADIUSD);
            return;
        }

        if(onlineCache.isExist(request.getAcctSessionId())){
            logger.error(request.getUsername(),"记账报文重复", Memarylogger.RADIUSD);
            return;
        }

        if (onlineCache.isLimitOver(user.getSubscriber(),user.getActiveNum())) {
            String errMessage = "用户<" + user.getSubscriber() + ">在线超过限制(MAX=" + user.getActiveNum() + ")(记账请求) " + request.getAcctStatusType();
            logger.error(request.getUsername(),errMessage, Memarylogger.RADIUSD);
            return;
        }

        RadiusOnline online = new RadiusOnline();
        online.setRadsec(false);
//        online.setId(CoderUtil.randomLongId());
        online.setNodeId(user.getNodeId());
        online.setRealname(user.getRealname());
        online.setUsername(request.getUsername());
        online.setNasId(request.getNasid());
        online.setNasAddr(nas.getIpaddr());
        online.setNasPaddr(request.getNasip());
        online.setSessionTimeout(request.getSessionTimeout());
        online.setFramedIpaddr(request.getFramedIPAddress());
        online.setFramedNetmask(request.getFramedIPNetmask());
        online.setMacAddr(request.getMacAddr());
        online.setInVlan(0);
        online.setOutVlan(0);
        online.setNasPort(0L);
        online.setNasClass("");
        online.setNasPortId(request.getNasPortId());
        online.setServiceType(0);
        online.setMacAddr(request.getMacAddr());
        online.setAcctSessionId(request.getAcctSessionId());
        online.setAcctSessionTime(request.getAcctSessionTime());
        online.setAcctInputTotal(request.getAcctInputTotal());
        online.setAcctOutputTotal(request.getAcctOutputTotal());
        online.setAcctInputPackets(request.getAcctInputPackets());
        online.setAcctOutputPackets(request.getAcctOutputPackets());
        online.setAcctStartTime(request.getAcctStartTime());
        onlineCache.putOnline(online);
        if(radiusConfig.isTrace()){
            logger.info(request.getUsername(),String.format(":: 新增用户在线信息: sessionId=%s", request.getAcctSessionId()), Memarylogger.RADIUSD);
        }
    }


    public void doStart(FreeradiusAcctRequest request, Bras nas, Subscribe user) throws RadiusException {
        addOnline(request,nas,user);
    }

    public void doUpdate(FreeradiusAcctRequest request, Bras nas, Subscribe user) throws RadiusException {
        if(!onlineCache.isExist(request.getAcctSessionId())){
            logger.info(request.getUsername(),":: 更新在线用户时用户在线记录不存在，立即新增在线信息 " + request.getUsername() + " ",Memarylogger.RADIUSD);
            addOnline(request,nas,user);
            return;
        }
        //2. 更新在线数据
        onlineCache.updateOnline(request);
    }


    public void doStop(FreeradiusAcctRequest request, Bras nas, Subscribe user) throws RadiusException {
        try {
            RadiusOnline online = onlineCache.removeOnline(request.getAcctSessionId());
            if(online==null){
                logger.info(request.getUsername(), ":: 收到记账下线消息(但在线用户不存在):" + request.toString(), Memarylogger.RADIUSD);
                return;
            }
            RadiusTicket radiusTicket = new RadiusTicket();
            radiusTicket.setId(CoderUtil.randomLongId());
            radiusTicket.setNodeId(online.getNodeId());
            radiusTicket.setUsername(online.getUsername());
            radiusTicket.setNasId(online.getNasId());
            radiusTicket.setNasAddr(online.getNasAddr());
            radiusTicket.setNasPaddr(online.getNasPaddr());
            radiusTicket.setAcctSessionTime(request.getAcctSessionTime());
            radiusTicket.setSessionTimeout(request.getSessionTimeout());
            radiusTicket.setFramedIpaddr(online.getFramedIpaddr());
            radiusTicket.setFramedNetmask(online.getFramedNetmask());
            radiusTicket.setMacAddr(online.getMacAddr());
            radiusTicket.setInVlan(online.getInVlan());
            radiusTicket.setOutVlan(online.getOutVlan());
            radiusTicket.setNasPort(online.getNasPort());
            radiusTicket.setNasClass(online.getNasClass());
            radiusTicket.setNasPortId(online.getNasPortId());
            radiusTicket.setNasPortType(0);
            radiusTicket.setServiceType(online.getServiceType());
            radiusTicket.setAcctSessionId(online.getAcctSessionId());
            radiusTicket.setAcctInputTotal(request.getAcctInputTotal());
            radiusTicket.setAcctOutputTotal(request.getAcctOutputTotal());
            radiusTicket.setAcctInputPackets(request.getAcctInputPackets());
            radiusTicket.setAcctOutputPackets(request.getAcctOutputPackets());
            radiusTicket.setAcctStopTime(new Date());
            radiusTicket.setAcctStartTime(DateTimeUtil.toDate(request.getAcctStartTime()));
            ticketCache.addTicket(radiusTicket);
        }catch (Exception e) {
            logger.error(request.getUsername(), "用户下线处理错误", e, Memarylogger.RADIUSD);
        }
    }

}
