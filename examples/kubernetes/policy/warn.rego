package main

import data.kubernetes


warn[msg] {
  kubernetes.is_service
  msg = "Services are not allowed"
}
