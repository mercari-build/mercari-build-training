# Tips for efficient development

## Reading and understanding error messages
When you read an error, it’s important to understand where the error has occurred so that you can use the appropriate keyword to search for the cause.
If there aren’t many hits for the word that you searched for, it’s highly possible that a part of the search term is wrong.

For example, have a look at the following error message:
```
docker: Cannot connect to the Docker daemon at unix:///var/run/docker.sock. 
Is the docker daemon running?.
See 'docker run --help'.
```
Even if your English skills are not strong, you should be able to assume from the error message, “Is the docker daemon running?” that something is wrong with the docker daemon.

Alternatively, looking at the error message,
the portion that reads “Cannot connect to the Docker daemon at unix:///var/run/docker.sock” could be a good keyword (or phrase) to search for.
When it comes to long error messages, it can be difficult to determine which part of the message to search for, but you will learn this over time. The text that comes after the word “error”, either at the very beginning or the very end of the message, tends to be very important, so be sure to read that part extra carefully.

## Ask questions at the right time and skillfully
If you read an error message and don’t understand it or you lose track of what the message means, ask those around you for help.

However, if you say something like,

> Step 4-1 doesn’t work.

it can be hard for the person you ask to grasp the situation and what you need to understand to fix it.

> At Step 4-1, I executed this command: ‘~~’
However, the following error occurred:
>
> Error message
>
> I think this error message means “**”.
> I therefore searched for “@@” which returned solutions X and Y. I attempted to apply these, but they did not fix the issue. Could you please help me out?

With this amount of information, the person you ask will be able to understand:
- Whether there is a problem with the environment
- Whether there is a problem with the project code
- Whether you have misunderstood something
- Whether you lack the knowledge needed to perform the task


And remember, you can ask anything, so don’t be shy!
If you search for an answer for 15 minutes and get no closer to a solution, talk to the other members on your team or to your mentor to see if they can give you a hint.
Solving problems by communicating with others is another kind of engineering skill.



## Read the official documentation
It’s of course best if you understand the official documentation (written in English), but it can be hard to find what you’re looking for in a document written in formal, technical English.
Even if you start by reading an article a third-party knowledge platform like [Qiita](https://qiita.com/) or Medium, etc. that describes a solution, once you understand the overall idea, make sure that you check the official documentation as well.

Example: To write Dockerfiles, see the [official reference materials](https://docs.docker.com/engine/reference/builder/).

Once you are able to do this, the development flow changes as follows:

Before

Copy a similar code, change it, and then try to make it work.   
-> Code does not work.   
-> Trial and error (This process uses up a lot of time.)

After

Start by understanding the problem (This takes a certain amount of time.)  
-> Understand and write

We often say “read the official documentation!” but if you don’t have the necessary basic or peripheral knowledge, reading the documentation won’t help you understand.
For example, for you to understand Docker, you need to have knowledge of Linux and networking.
To start off, use Udemy and relevant books to increase your basic computer science skills.


## Utilizing LLMs

- ChatGPT: https://chatgpt.com/
- DeepSeek: https://www.deepseek.com/
- Claude: https://claude.ai/
- Gemini: https://gemini.google.com/

LLMs (Large Language Models) like ChatGPT and Claude can be powerful development aids. However, there are several key points to keep in mind to effectively utilize LLMs.

### Ask Specific Questions

Rather than vague questions like "improve this code," ask specific questions like:

```
I'd like to reduce memory usage in the following Python code.
Particularly, I want to improve parts dealing with large arrays.
What methods are available?

[CODE]
```

### Share Complete Error Messages

When encountering errors, including the following information will help get more accurate responses:

1. Complete error message
2. Execution environment (OS and versions)
3. What you've tried and expected behavior

### Use as Code Review

Having LLMs review your code can help identify:

- Security concerns
- Performance improvements
- Better coding practices and design suggestions
- Comparisons with best practices

### Use as a Learning Tool

You can use LLMs as a "teacher" when learning new technologies or concepts:

```
Please explain Docker's layer architecture
in a way that beginners can understand.
I'd like the explanation to include concrete examples.
```

### Understand LLM Limitations

However, keep these points in mind:

- Always review and understand generated code before using it
- Double-check security-related aspects in official documentation
- Recognize that they may provide outdated or incorrect information
- **Don't share project confidential information**
  - [DeepSeek coding has the capability to transfer users' data directly to the Chinese government](https://abcnews.go.com/US/deepseek-coding-capability-transfer-users-data-directly-chinese/story?id=118465451)

LLMs are powerful tools, but they should be used as aids, with developers making the final decisions. Especially for beginners, it's recommended not to rely too heavily on LLMs and to properly learn the fundamentals independently.
