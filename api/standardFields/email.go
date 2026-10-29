package standardFields

//type EmailProcessor struct {
//	removePlusses, removePeriods bool
//}
//func (ep EmailProcessor) UnmarshalEmail(bs []byte) (em Email, err error){
//	json.Decoder{}
//	inp := string(bs)
//	ps := strings.Split(inp, "@")
//	if len(ps) != 2 {
//		return "", errors.New("invalid email format. Contains @ multiple times")
//	}
//	// Parse pre-domain
//	if ep.removePlusses {
//		ps[0] = strings.Split(ps[0], "+")[0]
//	}
//	if ep.removePeriods {
//		ps[0] = strings.Replace(ps[0], ".", "", -1)
//	}
//
//	// Parse domain
//	// TODO: this domain stuff
//
//	return Email(strings.Join(ps, "@")), nil
//}
//
//type Email string
//
//func (em *Email) UnmarshalJSON(bs []byte) (err error) {
//	inp := string(bs)
//	ps := strings.Split(inp, "@")
//	if len(ps) != 2 {
//		return errors.New("invalid email format. Contains @ multiple times")
//	}
//	// Parse pre-domain
//	// TODO: remove anything after plus symbol?
//	ps[0] = strings.Split(ps[0], "+")[0] // TODO: maybe remove?
//	// TODO: remove periods?
//	ps[0] = strings.Replace(ps[0], ".", "", -1) // TODO: maybe remove?
//
//	// Parse domain
//	// TODO: this domain stuff
//	*em = Email(strings.Join(ps, "@"))
//	return nil
//}
//
//// MarshalJSON is not custom.
////func (em Email) MarshalJSON() (bs []byte, err error) {
////	return []byte(em), nil
////}
