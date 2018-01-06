(define conz
  (lambda (a b)
    (lambda (c)
      (c a b))))
(define carr
  (lambda (c)
    (c (lambda (a b) a))))
(define cdrr
  (lambda (c)
    (c (lambda (a b) b))))

(define trick (conz 'mind 'blown))
(carr trick)
(cdrr trick)
trick
