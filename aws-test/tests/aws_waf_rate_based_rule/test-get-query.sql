select metric_name, partition, title
from aws.aws_waf_rate_based_rule
where rule_id = '{{ output.resource_id.value }}';
