package org.toughradius.mapper;

import org.toughradius.entity.Subscribe;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.springframework.stereotype.Repository;
import org.toughradius.form.SubscribeQuery;

import java.math.BigInteger;
import java.util.List;


@Repository
@Mapper
public interface SubscribeMapper {
	
	Subscribe findSubscribe(@Param(value = "username") String username);

	Subscribe findById(Long id);

	List<Subscribe> findLastUpdateUser(@Param(value = "lastUpdate") String lastUpdate);

	List<Subscribe> queryForList(SubscribeQuery subscribe);

	Integer queryTotal(SubscribeQuery subscribe);

	void insertSubscribe(Subscribe subscribe);

	void updateSubscribe(Subscribe subscribe);

	void deleteById(Long id);

	List<String> findAllUsername();

	Integer updateFlowAmountByUsername(@Param(value = "username") String username, @Param(value = "flowAmount") BigInteger flowAmount);

	 Integer updateMacAddr(@Param(value = "username") String username, @Param(value = "macAddr") String macAddr);

	 Integer updateInValn(@Param(value = "username") String username, @Param(value = "inValn") Integer inValn);

	 Integer updateOutValn(@Param(value = "username") String username, @Param(value = "outValn") Integer outValn);

	 void updatePassword(@Param(value = "id") Long id, @Param(value = "password") String password);

    void release(String ids);
}
