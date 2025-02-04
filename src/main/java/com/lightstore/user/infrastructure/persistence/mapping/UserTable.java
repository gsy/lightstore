package com.lightstore.user.infrastructure.persistence.mapping;


import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;
import lombok.ToString;

@NoArgsConstructor
@Getter
@Setter
@ToString
@Entity
@Table
public class UserTable {
    @Id
    private Integer id;
    private String userID;
    private String username;
}
