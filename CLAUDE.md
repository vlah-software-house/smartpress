# **AI Assistant Instructions & Project Context**

This file contains the core project specifications, technical stack, and strict workflow rules for any AI coding assistant or tool operating in this repository.  
**CRITICAL: You must read and adhere to these instructions before executing any commands, generating code, or modifying files.** As the project evolves, it is your responsibility to update this CLAUDE.md file to reflect new architectural decisions or workflow changes.

## **1\. Tech Stack**

* **Backend:** Go (Golang)  
* **Frontend:** HTMX, AlpineJS  
* **Database:** PostgreSQL  
* **Cache:** Valkey  
* **Testing:** Go standard testing (Unit/Functional), Playwright (E2E)  
* **Infrastructure:** Docker, Kubernetes

## **2\. Git & Development Workflow**

Follow this exact sequence for every new task, feature, fix, or patch:

### **Phase A: Planning & Branching**

1. **Confirm the Plan:** Before writing code, outline the plan and await user confirmation.  
2. **Branch Check:** Check the current branch. If it is NOT main, prompt the user to merge any outstanding changes into main and checkout main. Do not proceed until on main.  
3. **New Branch:** Once on main and the plan is confirmed, create and checkout a new, suggestively named branch (e.g., feat/add-user-auth, fix/cache-invalidation).  
4. **Stay on Branch:** Continue all work for the confirmed plan on this new branch until the user explicitly confirms the feature, fix, or patch is 100% complete.

### **Phase B: Step-by-Step Execution & Logging**

1. **Log Finished Steps:** For every confirmed finished step within the task, create a log entry file in the work/logs/ directory. Use a descriptive filename (e.g., work/logs/20260224\_implemented\_valkey\_connection.md).  
2. **Commit:** Immediately after writing the log file, commit the changes to the current branch with a clear, descriptive commit message.

### **Phase C: Completion & PR**

1. **Final Review:** Wait for the user to confirm the overall task is complete.  
2. **Push & PR:** Once confirmed, push the branch to the remote repository and create a Pull Request (or Merge Request).

## **3\. Go Coding Standards**

* **Version Check:** ALWAYS check the available Go version on the host machine (go version) before starting any code generation to ensure absolute compatibility.  
* **Best Practices:** Strictly adhere to idiomatic Go best practices, standard project layouts, and effective concurrency patterns.  
* **Readability:** All code must be heavily commented, clean, and easily readable by human developers. Document exported functions, structs, and complex logic blocks.

## **4\. Testing Requirements**

* **Coverage:** Implement tests everywhere. Aim for near 100% coverage.  
* **Unit & Functional:** Write comprehensive unit tests and functional tests for all Go backend logic.  
* **End-to-End (E2E):** Write E2E tests based on specification descriptions and user journeys in the interface. **You must use Playwright** for testing the HTMX/AlpineJS web pages.

## **5\. Infrastructure & DevOps**

* **Docker:** Use Docker for all local services (PostgreSQL, Valkey). Ensure a fully working Dockerfile and pipeline exist for building the application's main container.  
* **Kubernetes (K8s):**  
  * Deploy to the testing environment using Kubernetes manifests.  
  * Keep manifests updated, modular, and customizable for multiple environments (e.g., using Kustomize or Helm if appropriate, or structured manifest directories).  
* **Seeding Data:** Write and maintain database seeding scripts/manifests based on the application specifications to ensure a populated testing environment.  
* **Secrets Management:** Use a .secrets file for all environment variables provided by the user for prepared services in the testing environment. **Always ask the user for any missing variables** before attempting to run services. Do not hardcode secrets.
* **Database Management** Use the postgres user password to check existance of defined database name, and user that must be it's owner.

## **6\. AI Tool Coordination & Delegation**

* **Task Management:** Maintain a list of ongoing, pending, and completed tasks inside the work/tasks/ directory.  
* **Grouped Files:** Group tasks logically into files (e.g., work/tasks/frontend.md, work/tasks/database.md).  
* **Delegation:** Format these task files clearly so that *other* AI coding tools or agents can read them, understand the current project state, and pick up delegated work seamlessly.
