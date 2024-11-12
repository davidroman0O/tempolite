

=> Workflow
    - create workflow entity
    - create workflow execution

    if fail 
    if max attempt < attemps => retry
        - create workflow execution

        => Sub Workflow stepID
            - has hierarchy parent RunID + ParentID + StepID
                false
                    - create workflow entity
                    - create workflow execution
                true
                    - was entity completed? was execution completed?
                    true
                        - return output
                    false
                        - create workflow execution
                    
            if fail
            if max attempt < attemps => retry
                - create workflow execution
            else
                - workflow failed
    else
        - workflow failed



If restart
    - Entity: Running
    - Execution: Running
    then  
        Execution Running to Cancelled
        DO NOT INCREASE RETRY
        New Execution to Pending

