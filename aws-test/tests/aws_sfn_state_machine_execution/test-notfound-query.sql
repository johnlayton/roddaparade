select execution_arn, state_machine_arn, akas
from aws_sfn_state_machine_execution
where execution_arn = 'dummy-test-{{ output.execution_arn.value }}';
