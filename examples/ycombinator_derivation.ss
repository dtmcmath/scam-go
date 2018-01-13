(define add1 (lambda (x) (+ 1 x)))

(define eternity (lambda (x) (eternity x)))


;; Start with something that computes the length of empty lists

(define length0
  (lambda (l)
    (cond
     ([null? l) 0)
     (else (add1 (eternity (cdr l))))))
  )

(length0 '())
                                        ; ==> 0
; (length0 '(a))
                                        ; Never returns


(define length<=1
  (lambda (l)
    (cond
     ([null? l] 0)
     (else (add1 (length0 (cdr l)))))))
(define length<=1b
  (lambda (l)
    (cond
     ([null? l] 0)
     (else (add1
            (
             (lambda (l)
               (cond
                ([null? l) 0)
                (else (add1 (eternity (cdr l))))))
             (cdr l)
             )))))
  )

(define length<=1c
  (
   (lambda (length0)
     (lambda (l)
       (cond
        ((null? l) 0)
        (else (add1 (length0 (cdr l)))))))
   (lambda (l) ; this is length0
     (cond
      ((null? l) 0)
      (else (add1 (eternity (cdr ))))))
   )
  )

(length<=1c '(a))

(define length<=2
  (
   (lambda (length1)
     (lambda (l)
       (cond
        ((null? l) 0)
        (else (add1 (length1 (cdr l)))))))
   (
    (lambda (length0)
      (lambda (l)
        (cond
         ((null? l) 0)
         (else (add1 (length0 (cdr l)))))))
    length0
    )
   )
  )

(length<=2 '())
(length<=2 '(a))
(length<=2 '(a b))
    
(define length<=2b
  (
   (lambda (length0)
     ((lambda (gen-length)
        (gen-length (gen-length length0))
        )
      (lambda (length1)
        (lambda (l)
          (cond
           ((null? l) 0)
           (else (add1 (length1 (cdr l)))))))
      )
     )
   length0
   )
  )

(length<=2b '())
(length<=2b '(a))
(length<=2b '(a b))

;;
;; Using "mk-length"

(define length0
  (
   (lambda (mk-length)
     (mk-length eternity))
   (lambda (le)
     (lambda (l)
       (cond
        ((null? l) 0)
        (else (add1 (le (cdr l))))))
     )
   )
  )

(length0 '())

(define length<=1
  (
   (lambda (mk-length)
     (mk-length
      (mk-length eternity)))
   (lambda (le)
     (lambda (l)
       (cond
        ((null? l) 0)
        (else (add1 (le (cdr l))))))
     )
   )
  )

(length<=1 '())
(length<=1 '(a))

(define length<=2
  (
   (lambda (mk-length)
     (mk-length
      (mk-length
       (mk-length eternity))))
   (lambda (le)
     (lambda (l)
       (cond
        ((null? l) 0)
        (else (add1 (le (cdr l))))))
     )
   )
  )

(length<=2 '())
(length<=2 '(a))
(length<=2 '(a b))

(define length<=3
  (
   (lambda (mk-length)
     (mk-length
      (mk-length
       (mk-length
        (mk-length eternity)))))
   (lambda (le)
     (lambda (l)
       (cond
        ((null? l) 0)
        (else (add1 (le (cdr l))))))
     )
   )
  )

(length<=3 '())
(length<=3 '(a))
(length<=3 '(a b))
(length<=3 '(a b c))

;;
;; Without eternity

(define length0
  (
   (lambda (mk-length)
     (mk-length mk-length))
   (lambda (le)
     (lambda (l)
       (cond
        ((null? l) 0)
        (else (add1 (le (cdr l))))))
     )
   )
  )

(length0 '())

(define length<=1
  (
   (lambda (mk-length)
     (mk-length mk-length))
   (lambda (mk-length)
     (lambda (l)
       (cond
        ((null? l) 0)
        (else (add1 ((mk-length eternity) (cdr l))))))
     )
   )
  )

(length<=1 '())
(length<=1 '(a))

(define ur-length
  (
   (lambda (mk-length)
     (mk-length mk-length))
   (lambda (mk-length)
     (lambda (l)
       (cond
        ((null? l) 0)
        (else (add1 ((mk-length mk-length) (cdr l))))))
     )
   )
  )

(ur-length '())
(ur-length '(a))
(ur-length '(a b))
(ur-length '(a b c))
(ur-length '(a b c d))
(ur-length '(a b c d e))
(ur-length '(a b c d e f))

;; (define never-length
;;   (
;;    (lambda (mk-length)
;;      (mk-length mk-length))
;;    (lambda (mk-length)
;;      ((lambda (le)
;;         (lambda (l)
;;           (cond
;;            ((null? l) 0)
;;            (else (add1 (le (cdr l))))))
;;         )
;;       (mk-length mk-length)
;;       )
;;      )
;;    )
;;   )


(define near-length
  (
   (lambda (mk-length)
     (mk-length mk-length))
   (lambda (mk-length)
     ((lambda (length)
        (lambda (l)
          (cond
           ((null? l) 0)
           (else (add1 (length (cdr l)
                               ))))
          )
        )
      (lambda (x) ((mk-length mk-length) x))
      )
     )
   )
  )

(near-length '(a b c d e f))
(near-length '(a b c d e))
(near-length '(a b c d))
(near-length '(a b c))
(near-length '(a b))
(near-length '(a))
(near-length '())


(define Y-length
  ((lambda (le)
     (
      (lambda (mk-length)
        (mk-length mk-length))
      (lambda (mk-length)
        (
         le       
         (lambda (x) ((mk-length mk-length) x))
         )
        )
      )
     )
   (lambda (length) ; this is le
     (lambda (l)
       (cond
        ((null? l) 0)
        (else (add1 (length (cdr l)
                            ))))
       )
     )
   )
  )

(Y-length '(a b c))

(define Y
  (lambda (le)
    (
     (lambda (f) (f f))
     (lambda (g)
       (le (lambda (x) ((g g) x))))
     )
    )
  )

(define length-Y
  (Y (lambda (length)
       (lambda (l)
         (cond
          ((null? l) 0)
          (else (add1 (length (cdr l)))))))))


(length-Y '(a b c d e f))
(length-Y '(a b c d e))
(length-Y '(a b c d))
(length-Y '(a b c))
(length-Y '(a b))
(length-Y '(a))
(length-Y '())
