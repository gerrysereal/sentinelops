terraform {
  required_version = ">= 1.6.0"
}

variable "cluster_name" {
  type        = string
  description = "Target cluster name."
  default     = "sentinelops-k3s"
}

output "cluster_name" {
  value = var.cluster_name
}
