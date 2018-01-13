; Chapter 9

(define add1 (lambda (n) (+ 1 n)))

(define Y
  (lambda (le)
    ((lambda (f) (f f))
     (lambda (f)
       (le (lambda (x) ((f f) x)))))))

(define length-without-recursion
  (Y (lambda (length)
       (lambda (l)
         (cond ([null? l] 0)
               (else (add1 (length (cdr l)))))))))

(length-without-recursion '(a b c d e))
; ==> 5

(define rember-without-recursion
  (lambda (a l)
    ((Y (lambda (rember-a)
          (lambda (l)
            (cond ([null? l] '())
                  ([eq? a (car l)] (cdr l))
                  (else (cons (car l) (rember-a (cdr l))))))))
     l)))

(rember-without-recursion 'a '(a b c d))
; ==> '(b c d)
(rember-without-recursion 'b '(a b c d))
; ==> '(a c d)
(rember-without-recursion 'c '(a b c d))
; ==> '(a b d)
(rember-without-recursion 'd '(a b c d))
; ==> '(a b c)
