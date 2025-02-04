package com.lightstore.user.infrastructure.persistence.repository;

import com.lightstore.user.infrastructure.persistence.mapping.UserTable;
import org.springframework.data.repository.Repository;

public interface JdbcUserRepository extends Repository<UserTable, String> {
    UserTable save(UserTable user);
}
