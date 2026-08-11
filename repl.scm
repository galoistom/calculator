(display "type 'exit to exit")
(define (fold func init l)
  (if (null? l)
      init
      (fold func (func init (car l)) (cdr l))))
(define (foldr func init l)
  (if (null? l)
      init
      (func (car l) (foldr func init (cdr l)))))
(let loop ()
  (let ((x (readline)))
    (if (eq? x 'exit)
        (display "good bye!")
        (loop))))
