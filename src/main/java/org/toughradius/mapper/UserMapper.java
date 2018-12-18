package org.toughradius.mapper;

import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.springframework.stereotype.Repository;
import org.toughradius.entity.User;

import java.math.BigInteger;
import java.util.List;


@Repository
@Mapper
public interface UserMapper {

    BigInteger getFlowAmount(String username);

    User findUser(@Param(value = "username") String username);

    List<User> findLastUpdateUser(@Param(value = "lastUpdate") String lastUpdate);

    int updateFlowAmountByUsername(@Param(value = "username") String username, @Param(value = "flowAmount") Long flowAmount);

    int updateMacAddr(@Param(value = "username") String username, @Param(value = "macAddr") String macAddr);

    int updateInValn(@Param(value = "username") String username, @Param(value = "inValn") Integer inValn);

    int updateOutValn(@Param(value = "username") String username, @Param(value = "outValn") Integer outValn);
}
