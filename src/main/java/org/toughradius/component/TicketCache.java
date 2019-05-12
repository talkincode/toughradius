package org.toughradius.component;

import org.toughradius.common.DateTimeUtil;
import org.toughradius.common.PageResult;
import org.toughradius.common.ValidateUtil;
import org.toughradius.config.RadiusConfig;
import org.toughradius.entity.RadiusTicket;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor;
import org.springframework.stereotype.Component;

import java.io.*;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ConcurrentLinkedDeque;
import java.util.zip.GZIPInputStream;
import java.util.zip.GZIPOutputStream;

@Component
public class TicketCache {

    private Log logger = LogFactory.getLog(TicketCache.class);
    private final static ConcurrentLinkedDeque<RadiusTicket> queue = new  ConcurrentLinkedDeque<>();

    @Autowired
    private RadiusConfig radiusConfig;

    public void addTicket(RadiusTicket ticket)
    {
        queue.addFirst(ticket);
    }

    public void syncData(){
        try {
            List<RadiusTicket> logs = new ArrayList<RadiusTicket>();
            int count = 0;
            while(queue.size() > 0 && count <= 4096){
                logs.add(queue.removeFirst());
                count++;
            }
            File logdir = new File(radiusConfig.getTicketDir());
            if(!logdir.exists()){
                logdir.mkdirs();
            }
            BufferedOutputStream out = null;
            try {
                String filename = String.format("%s/radius-ticket.%s.txt",radiusConfig.getTicketDir(), DateTimeUtil.getDateString());
                File tfile = new File(filename);
                boolean isnew = !tfile.exists();
                out = new BufferedOutputStream(new FileOutputStream(tfile, true));
                if(isnew){
                    out.write(RadiusTicket.getHeaderString().getBytes("utf-8"));
                    out.write("\n".getBytes());
                }
                for(RadiusTicket ticket : logs){
                    out.write(ticket.toString().getBytes("utf-8"));
                    out.write("\n".getBytes());
                }
            } catch (Exception e) {
                logger.error("上网日志写入出错",e);
            } finally {
                try {
                    if (out != null) {
                        out.close();
                    }
                } catch (IOException e) {
                    e.printStackTrace();
                }
            }
        } catch (Exception e) {
            logger.error("Sync ticket error:",e);
        }
    }

    public PageResult<RadiusTicket> queryTicket(int start,
                                                int count,
                                                String startDate,
                                                String endDate,
                                                String nasid,
                                                String nasaddr,
                                                Integer nodeId,
                                                String username,
                                                String keyword) throws ServiceException {
        int rowNum = 0;
        if(ValidateUtil.isEmpty(startDate)){
            startDate = DateTimeUtil.getDateString()+" 00:00:00";
        }
        if(ValidateUtil.isEmpty(endDate)){
            endDate = DateTimeUtil.getDateString()+" 23:59:59";
        }

        if(startDate.length() == 16){
            startDate += ":00";
        }

        if(endDate.length() == 16){
            endDate += ":59";
        }

        if(start + count > 10000){
            throw new ServiceException("查询最大数量为10000");
        }
        if(!ValidateUtil.isDateTime(startDate)){
            throw new ServiceException("查询开始时间格式必须许为 yyyy-MM-dd HH:mm:ss");
        }
        if(!ValidateUtil.isDateTime(endDate)){
            throw new ServiceException("查询开始时间格式必须许为 yyyy-MM-dd HH:mm:ss");
        }

        if(DateTimeUtil.compareSecond(endDate,startDate)>(85400*30)){
            throw new ServiceException("查询时间跨度不能超过30天");
        }

        BufferedReader reader = null;
        String beginDay = startDate.substring(0, 10);
        String endDay = endDate.substring(0, 10);
        try
        {
            String filename = String.format("%s/radius-ticket",radiusConfig.getTicketDir());
            String currendDate = DateTimeUtil.getDateString();
            int index = 0, end = start + count;
            ArrayList<RadiusTicket> list = new ArrayList<RadiusTicket>();

            boolean loop = true;
            while (beginDay.compareTo(endDay) <= 0 && loop)
            {
                String curFileName = String.format("%s.%s.txt" , filename,beginDay);

                File file = new File(curFileName);
                if (!file.exists())
                {
                    beginDay = DateTimeUtil.getNextDateString(beginDay);
                    continue;
                }

                reader = new BufferedReader(new InputStreamReader(new FileInputStream(file), "UTF-8"));
                String line = null;
                while ((line = reader.readLine()) != null)
                {
                    RadiusTicket logdata = RadiusTicket.fromString(line);
                    if(logdata==null){
                        continue;
                    }

                    if (ValidateUtil.isNotEmpty(username) && !logdata.getUsername().contains(username))
                        continue;

                    if (ValidateUtil.isNotEmpty(nasid) && !nasid.equalsIgnoreCase(logdata.getNasId()))
                        continue;

                    if (ValidateUtil.isNotEmpty(nasaddr) && !nasaddr.equalsIgnoreCase(logdata.getNasAddr()))
                        continue;

                    if (nodeId!=null && nodeId != logdata.getNodeId().intValue())
                        continue;

                    if (ValidateUtil.isNotEmpty(keyword) &&
                            ( !logdata.getUsername().contains(keyword)
                                    && !logdata.getFramedIpaddr().contains(keyword)
                                    && !logdata.getMacAddr().contains(keyword) ))
                        continue;

                    if(DateTimeUtil.compareSecond(logdata.getAcctStartTime(),DateTimeUtil.toDate(startDate))<0){
                        continue;
                    }
                    if(DateTimeUtil.compareSecond(logdata.getAcctStartTime(),DateTimeUtil.toDate(endDate))>0){
                        continue;
                    }

                    index++;
                    if (index >= start && index <= end){
                        if(list.size()<count){
                            list.add(logdata);
                        }
                    }
                    rowNum ++;
                    if(rowNum>=10000){
                        loop = false;
                        break;
                    }
                }

                beginDay = DateTimeUtil.getNextDateString(beginDay);
                reader.close();
            }
            return new PageResult<RadiusTicket>(start, rowNum, list);
        }
        catch (Exception e)
        {
            e.printStackTrace();
            throw new ServiceException("查询上网记录失败",e);
        }
        finally
        {
            if (reader != null)
                try{reader.close();}catch(Exception e){}

            reader = null;
        }

    }





}
