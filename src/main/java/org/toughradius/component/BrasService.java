package org.toughradius.component;

import org.toughradius.common.ValidateUtil;
import org.toughradius.entity.Bras;
import org.toughradius.mapper.BrasMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class BrasService {

	@Autowired
	private BrasMapper brasMapper;

	@Autowired
	private Memarylogger logger;

	/**
	 * 查找 BRAS 信息
	 * @param ipaddr 数据包来源IP
	 * @param srcip 设备中配置的IP，可能是内网IP
	 * @param identifier  设备唯一标识
	 * @return
	 * @throws ServiceException
	 */
	public Bras findBras(String ipaddr, String srcip, String identifier) throws ServiceException{
		Bras tcBras = null;
		if(ValidateUtil.isNotEmpty(ipaddr)&&!"0.0.0.0".equals(ipaddr)){
			tcBras = brasMapper.findByIPAddr(ipaddr);
		}

		if(tcBras == null && ValidateUtil.isNotEmpty(srcip)&&!"0.0.0.0".equals(srcip)){
			tcBras = brasMapper.findByIPAddr(srcip);
		}

		if (tcBras == null && ValidateUtil.isNotEmpty(identifier)) {
			tcBras = brasMapper.findByidentifier(identifier);
		}

		if (tcBras == null) {
			String message = String.format("Bras设备 id=%s, ip=%s 不存在", identifier, ipaddr);
			logger.error(message, Memarylogger.RADIUSD);
			throw new ServiceException(message);
		}

		if (tcBras.getStatus() != null && "disabled".equals(tcBras.getStatus())) {
			String message = String.format("Bras设备 id=%s, ip=%s 已停用", identifier, ipaddr);
			logger.error(message, Memarylogger.RADIUSD);
			throw new ServiceException(message);
		}

		return tcBras;
	}

	public List<Bras> queryForList(Bras bras){
		return brasMapper.queryForList(bras);
	}

	public void insertBras(Bras bras){
		brasMapper.insertBras(bras);
	}

	public void updateBras(Bras bras){
		brasMapper.updateBras(bras);
	}

	public void deleteById(Long id){
		brasMapper.deleteById(id);
	}

	public Bras selectByidentifier(String identifier){
		return brasMapper.findByidentifier(identifier);
	}

	public Bras selectByIPAddr(String ipaddr){
		return brasMapper.findByIPAddr(ipaddr);
	}

	public Bras selectById(Long id){
		return brasMapper.findById(id);
	}
}
