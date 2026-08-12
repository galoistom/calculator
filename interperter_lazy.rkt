(define (my-eval exp env)
  (cond ((self-evaluating? exp) exp)
        ((variable? exp) (lookup-variable-value exp env))
        ((quoted? exp) (text-of-quotation exp))
        ((assignment? exp) (eval-assignment exp env))
        ((definition? exp) (eval-definition exp env))
        ((let? exp) (eval-let (cdr exp) env))
        ((if? exp) (eval-if exp env))
        ((and? exp) (eval-and (cdr exp) env))
        ((or? exp) (eval-or (cdr exp) env))
        ((lambda? exp)
         (make-procedure (cadr exp) (cddr exp) env))
        ((begin? exp)
         (eval-sequence (begin-actions exp) env))
        ((cond? exp)
         (my-eval (cond->if exp) env))
        ((application? exp)
         (my-apply (actual-value (car exp) env)
                (cdr exp) env))
        (else
         (error "Unknown expression type - EVAL" exp))))

(define (my-apply procedure arguments env)
  (cond ((primitive-procedure? procedure)
         (apply-primitive-procedure
          procedure (list-of-arg-values arguments env)))
        ((compound-procedure? procedure)
         (eval-sequence
          (procedure-body procedure)
          (extend-environment
           (procedure-parameters procedure)
           (list-of-delayed-args arguments env)
           (procedure-environment procedure))))
        (else
         (error "Unknown procedure type - APPLY" procedure))))

(define (delay-it exp env) (list 'thunk exp env))

(define (list-of-delayed-args exps env)
  (if (null? exps)
      '()
      (cons (delay-it (car exps) env)
            (list-of-delayed-args (cdr exps) env))))

(define (list-of-arg-values exps env)
  (if (null? exps)
      '()
      (cons (actual-value (car exps) env)
            (list-of-arg-values (cdr exps) env))))

(define (thunk? obj) (tagged-list? obj 'thunk))

(define (evaluated-thunk? obj)
  (tagged-list? obj 'evaluated-thunk))

(define (force-it obj)
  (cond ((thunk? obj)
         (let ((result (actual-value (cadr obj) (caddr obj))))
           (set! obj (list 'evaluated-thunk result '()))
           result))
        ((evaluated-thunk? obj)
         (cadr obj))
        (else obj)))

(define (actual-value exp env)
  (force-it (my-eval exp env)))

(define (eval-and exps env)
  (define (iter exp)
    (if (null? exp) 'true
        (if (eval (car exp) env)
            (iter (cdr exp))
            'false)))
  (iter exps))

(define (eval-or exps env)
  (define (iter exp)
    (if (null? exp) 'false
        (if (eval (car exp) env)
            'true
            (iter (cdr exp)))))
  (iter exps))

(define (and? exp) (tagged-list? exp 'and))

(define (or? exp) (tagged-list? exp 'or))

(define (eval-let exp env)
  (if (symbol? (car exp))
      (let* ((name (car exp))
             (assignments (cadr exp))
             (actions (caddr exp))
             (variables (map car assignments))
             (values (map (lambda (x) (my-eval (cadr x) env)) assignments))
             (function (list 'define (cons name variables) actions))
             (procedures (sequence->exp (list function (cons name values)))))
        (eval (list 'begin procedures) env))
      (let* ((assignments (car exp))
             (actions (cdr exp))
             (variables (map car assignments))
             (values (map (lambda (x) (my-eval (cadr x) env)) assignments))
             (new-env (extend-environment variables values env)))
        (eval (sequence->exp actions) new-env))))

(define (let? exp) (tagged-list? exp 'let))

(define (list-of-values exps env)
  (if (null? exps)
      '()
      (cons (eval (car exps) env)
            (list-of-values (cdr exps) env))))

(define (application? exp) (pair? exp))

(define (cond->if exp)
  (expand-clauses (cdr exp)))
(define (expand-clauses clauses)
  (if (null? clauses)
      'false
      (let ((first (car clauses))
            (rest (cdr clauses)))
        (cond ((eq? (car first) 'else)
               (if (null? rest)
                   (sequence->exp (cdr first))
                   (error "ELSE claus isn't lase - COND->IF")))
              ((and (pair? (cdr first))
                    (pair? (cddr first))
                    (eq? (cadr first) '=>))
               `(let ((temp ,(car first)))
                  (if temp (,(caddr first) temp)
                      ,(expand-clauses rest))))
              (else (make-if (car first)
                             (sequence->exp (cdr first))
                             (expand-clauses rest)))))))

(define (cond? exp) (tagged-list? exp 'cond))

(define (begin-actions exp) (cdr exp))

(define (sequence->exp seq)
  (cond ((null? seq) seq)
        ((null? (cdr seq)) (car seq))
        (else (cons 'begin seq))))

(define (eval-sequence exp env)
  (cond ((null? (cdr exp)) (my-eval (car exp) env))
        (else (my-eval (car exp) env)
              (eval-sequence (cdr exp) env))))

(define (begin? exp) (tagged-list? exp 'begin))

(define (make-procedure parameters body env)
  (list 'procedure (list parameters body) env))

(define (lambda? exp) (tagged-list? exp 'lambda))

(define (eval-if exp env)
  (if (actual-value (cadr exp) env)
      (my-eval (caddr exp) env)
      (my-eval (if (null? (cdddr exp)) false (cadddr exp)) env)))

(define (make-if pre con alt)
  (list 'if pre con alt))

(define (if? exp) (tagged-list? exp 'if))
(define (eval-definition exp env)
  (define-variable! (definition-valriable exp)
    (my-eval (definition-value exp) env)
    env)
  'ok)
(define (definition-value exp)
  (if (symbol? (cadr exp))
      (caddr exp)
      (make-lambda (cdadr exp) (cddr exp))))

(define (make-lambda par body)
  (cons 'lambda (cons par body)))

(define (definition? exp) (tagged-list? exp 'define))

(define (eval-assignment exp env)
  (set-valiable-value! (car exp) (my-eval (cadr exp) env) env))

(define (assignment? exp) (tagged-list? exp 'set!))

(define (tagged-list? exp tag)
  (if (pair? exp)
      (eq? (car exp) tag)
      false))

(define (text-of-quotation exp) (cadr exp))

(define (quoted? exp)
  (tagged-list? exp 'quote))

(define (variable? exp)
  (symbol? exp))

(define (self-evaluating? exp)
  (cond ((number? exp) true)
        ((string? exp) true)
        (else false)))

(define (definition-valriable exp)
  (if (symbol? (cadr exp))
      (cadr exp)
      (caadr exp)))

(define (define-variable! var val env)
  (hash-ref! (car env) var val))

(define the-empty-environment '())

(define (make-frame variables values)
  (let ((h (make-hash)))
    (for-each (lambda (k v) (hash-set! h k v)) variables values)
    h))

(define (add-bindings-to-frame! var val frame)
  (hash-set! frame var val))

(define (extend-environment vars vals base-env)
  (if (= (length vars) (length vals))
      (cons (make-frame vars vals) base-env)
      (if (< (length vars) (length vals))
          (error "Too many arguments" vars vals)
          (error "Too few arguments" vars vals))))

(define (lookup-variable-value var env)
  (define (env-loop env)
    (if (null? env)
        (error "Unbounded variable" var)
        (let ((val (hash-ref (car env) var)))
          (if (null? val)
              (env-loop (cdr env))
              val))))
  (env-loop env))

(define (set-valiable-value! var val env)
  (define (env-loop env)
    (cond
      ((null? env)(error "Unbounded variable" var))
      ((hash-has-key? (car env) var) (hash-set! (car env) var val))
      (else (env-loop (cdr env)))))
  (env-loop env))

;; (define apply-in-underlying-scheme 'apply)


(define (comound-procedure? p)
  (tagged-list? p 'procdure))

(define (primitive-procedure? proc) (tagged-list? proc 'primitive))

(define (apply-primitive-procedure proc args)
  (apply (cadr proc) args))

(define (setup-environment)
  (let ((initial-env (extend-environment (map car primitive-procedure)
                                      (primitive-procedure-objects)
                                      the-empty-environment)))
    (define-variable! 'true true initial-env)
    (define-variable! 'false false initial-env)
    initial-env))

(define (primitive-procedure-objects)
  (map (lambda (proc) (list 'primitive (cadr proc)))
       primitive-procedure))

(define primitive-procedure
 (list (list 'car car)
        (list 'cdr cdr)
        (list 'cadr cadr)
        (list 'cons cons)
        (list 'null? null?)
        (list 'not not)
        (list '+ +)
        (list '- -)
        (list '* *)
        (list '/ /)
        (list '= =)
        (list '< <)
        (list '> >)
        (list 'list list)
        (list 'eq? eq?)
        (list 'display display)
        (list 'newline newline)
        (list 'exit exit)))


(define (driver-loop)
  (let ((input (readline-raw)))
    (let ((output (my-eval input the-global-environment)))
      (user-print output)))
  (driver-loop))

(define (compound-procedure? p)
  (tagged-list? p 'procedure))

(define (procedure-parameters p) (caadr p))

(define (procedure-body p) (cadadr p))

(define (procedure-environment p) (caddr p))

(define (user-print object)
  (if (compound-procedure? object)
      (display (list 'compound-procedure
                     (procedure-parameters object)
                     (procedure-body object)
                     '<procedure-env>))
      (display object)))

(define the-global-environment (setup-environment))

(display "a small scheme interpreter running on scheme")
;; (display the-global-environment)
(driver-loop)
