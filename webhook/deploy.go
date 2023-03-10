package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

type Deploy struct{}

func (d Deploy) Validating(w http.ResponseWriter, r *http.Request) {

	// some check
	if r.Header.Get("Content-Type") != "application/json" {
		sendError(fmt.Errorf("request content-type=%s, not equal application/json", r.Header.Get("Content-Type")), w)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendError(err, w)
		return
	}

	//
	var ar admissionv1.AdmissionReview
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		sendError(err, w)
		return
	}

	if ar.Request == nil {
		sendError(fmt.Errorf("ar.Request == nil"), w)
		return
	}

	// ar response handler logic in this func
	// create a new adminssionReview for response
	reviewRes := d.Deployment(&ar)
	resReview := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ar.APIVersion,
			Kind:       ar.Kind,
		},
	}
	resReview.Response = reviewRes

	// rewrite resReview back to webhook response
	responBody, err := json.Marshal(resReview)
	if err != nil {
		sendError(err, w)
		return
	}

	// println res Body to check status
	// log.Printf("response body = %s", string(responBody))
	if _, err := w.Write(responBody); err != nil {
		sendError(err, w)
		return
	}
	log.Printf("Exec validating webhook success")
}

func sendError(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "%s", err)
}

func (d Deploy) Deployment(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	// try to get deployment object, and modify it
	var deploy *appsv1.Deployment
	if err := json.Unmarshal(ar.Request.Object.Raw, &deploy); err != nil {
		log.Printf("Unmarshal ar.Request.Object.Raw to deploy err:%v\n", err)
		ar.Response = &admissionv1.AdmissionResponse{
			Allowed: false,
			UID:     ar.Request.UID,
			Result: &metav1.Status{
				Message: `Unmarshal ar.Request.Object.Raw to deploy err`,
				Code:    514,
				Status:  "Failure",
			},
		}
		return ar.Response
	}

	// check logical
	const CANARY string = "canary"
	const CICD_ENV string = "cicd_env"
	nameHasCanary, _ := regexp.MatchString(CANARY, deploy.Name)
	if v, ok := deploy.Spec.Template.ObjectMeta.Labels[CICD_ENV]; ok && !nameHasCanary {
		if v == CANARY {
			log.Printf(`Not allowed to apply deployment=%s.%s, name not match "canary" and labels had "%s: %s"`,
				deploy.Namespace, deploy.Name, CICD_ENV, CANARY)
			ar.Response = &admissionv1.AdmissionResponse{
				Allowed: false,
				UID:     ar.Request.UID,
				Result: &metav1.Status{
					Message: fmt.Sprintf(`Not allowed to apply deployment=%s.%s, name not match "canary" and labels had "%s: %s"`,
						deploy.Namespace, deploy.Name, CICD_ENV, CANARY),
					Code:   414,
					Status: "Failure",
				},
			}

			return ar.Response

		}

	}

	// default check ok response
	log.Printf("deployment %s/%s pass through validating webhook", deploy.Namespace, deploy.Name)
	ar.Response = &admissionv1.AdmissionResponse{
		Allowed: true,
		UID:     ar.Request.UID,
		Result: &metav1.Status{
			Message: fmt.Sprintf(`Allowed to apply deployment=%s.%s, check name and label ok`, deploy.Namespace, deploy.Name),
			Status:  "Success",
		},
	}

	return ar.Response
}
