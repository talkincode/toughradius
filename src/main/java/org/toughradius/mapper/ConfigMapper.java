package org.toughradius.mapper;

import org.toughradius.entity.Config;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.springframework.stereotype.Repository;

import java.util.List;


@Repository
@Mapper
public interface ConfigMapper {

    Integer getInterimTimes();

    Integer getIsCheckPwd();

    void insertConfig(Config config);

    void updateConfig(Config config);

    void deleteById(Integer id);

    void deleteConfig(@Param(value = "type") String type, @Param(value = "name") String name);

    Config findConfig(@Param(value = "type") String type, @Param(value = "name") String name);

    List<Config> queryForList(@Param(value = "type") String type);

}