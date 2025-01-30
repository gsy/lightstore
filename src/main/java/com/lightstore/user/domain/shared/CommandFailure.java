package com.lightstore.user.domain.shared;

import lombok.AllArgsConstructor;

import java.util.Set;

@AllArgsConstructor
public class CommandFailure {
    public final Set<Integer> codes;
}
