**==Work in progress==**

**About**

Chainrunner is a lightweight utility for running command chains on local and remote hosts. This can be used as an assistant in ci / cd or for performing routine tasks.

**Roadmap**
 - Add remote host commands support
 - Write Documentation and Examples
 - Add supporting of running few tasks parallel
 - Add template variables in chain providers
   - source
        ```yaml
        root:
         - command {{variable}}
        ```
   - command
        ```shell script
        chainrun -o variable=value --chainFile source.yaml
        ```
   - result
       ```yaml
       root:
        - command value
       ```
 - Add environment variables support