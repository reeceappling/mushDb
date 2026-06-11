'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {
    AllEntries,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {PlateData} from "@/app/components/plateServer";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {InnocDisplay, TransfersOutDisplay} from "@/app/components/transferClient";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import {
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoUpdateMultipartRequest,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    ImportDisplayInput,
    ImportEntryFormWrapper,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    MultipartImportRequest,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey, RequiredKey,
    resolveContamsFormData,
    resolvePicsFormData,
    setFormData,
    setFormImages,
    YesNoSelector,
} from "@/app/components/common";
import ReaderWriterSelector, {
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {
    ErrorDisplay,
    GensFormDisplay,
    MostRecentImageDisplay,
    ParentDisplay,
    PicsDisplay
} from "@/app/components/formSubcomponents/commonClient";
import {
    AgarBatchArea,
} from "@/app/components/agarBatchClient";
import {
    ContaminationForm,
    ContamsDisplay,
    InitialContamState,
    IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import {AgarBatchData, AgarBatchSelectorCloseable} from "@/app/components/agarBatchServer";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SaleArea} from "@/app/components/saleClient";
import {
    ExistingSpeciesSelector,
    ExistingSpeciesSubspeciesSelector,
    SpeciesSubspeciesArea
} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {
    AclDisplay,
    AssertACL,
    IsValidAcl,
    MarshalAcl,
    TogglableAreaWithDepth, UnmarshalAcl
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {InputNumberWithSmallTitle} from "@/app/components/formSubcomponents/numericInput";
import Box from "@mui/material/Box";
import Slider from "@mui/material/Slider";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import TestAndValidate from "@/app/components/testing/untested";

export function AssertPlate(input: any): asserts input is PlateData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Plate assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['agarBatch', 'string'],
        ['species', 'string'],
        ['subspecies', 'string'],
        ['innoc', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['knownFruitable', 'boolean'],
        ['sale', 'string'],
        ['disposed', 'number'],
        ['condensationCoverageAtSealTime', 'number'],
        ['pourCoverage', 'number'],
        ['wetAtCooledTime', 'boolean'],
        ['agarOnOutsideAtPourTime', 'boolean'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Plate assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl],
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Plate assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    const complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
    ])
    for (const [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Plate assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['pics', IsValidPicWithNotesIncoming],
        ['contamination', IsValidContamination],
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Plate assertion failure: optional array key ' + key + ' was not valid');
        }
    }

    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export function PourCoverageSelector({value, setPourCoverage}: {
    value?: number,
    setPourCoverage: (value?: number) => void,
}) {
    return <OptionalSliderSelector txt={"Pour Coverage"} label={"Pour Coverage (%)"} initial={value} min={0} def={100}
                                   max={100} updateParent={setPourCoverage}/>
}
export function PourCoverageSelectorRequired({value, setPourCoverage}: {
    value?: number,
    setPourCoverage: (value?: number) => void,
}) {
    return <NumberSlider defaultValue={100} min={0} max={100} label={"pour coverage"}
                         onChange={(e, v, a) => {
                             e.stopPropagation();
                             setPourCoverage(v)
                         }}/>
}

export default function PlateDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput<PlateData>) {
    const [initial, setInitial] = useState(data as PlateData)

    const [images, setImages] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(data.pics))
    const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(data.contamination))
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(data.knownFruitable)
    const [pourCoveragePct, setPourCoveragePct] = useState(initial.pourCoverage)
    const [condensationCoveragePct, setCondensationCoveragePct] = useState(initial.condensationCoverageAtSealTime)
    const [wetAtCooledTime, setWetAtCoolTime] = useState(initial.wetAtCooledTime)
    const [agarOnOutsideAtPourTime, setAgarOnOutsideAtPourTime] = useState(initial.agarOnOutsideAtPourTime)
    const [sale, setSale] = useState<string | undefined>(data.sale)
    const [disposed, setDisposed] = useState<number | undefined>(data.disposed)
    const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
    const [acl, setAcl] = useState<ACL>(initial.acl)
    // Helper states
    const [transfersOut, setTransfersOut] = useState<string[]>(data.transfersOut || [])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const updateInitial = (updated: PlateData) => {
        setInitial(updated)
        setImages(InitialPicsEntries(updated.pics))
        setContams(InitialContamState(updated.contamination))
        setKnownFruitable(updated.knownFruitable)
        setSale(updated.sale)
        setDisposed(updated.disposed)
        setNotes(InitialNotesState(updated.notes))
        setTransfersOut(updated.transfersOut || [])
        setAcl(updated.acl)
        setErr(undefined)
    }
    const cookies = useContext(CookiesContext)
    const submit = () => {
        console.log("creating update request")
        const formData = new FormData()
        const dataObj: any = {
            knownFruitable: knownFruitable,
            sale: sale,
            disposed: disposed,
            notes: notes,
            writeTagTo: writeTagTo,
            acl: MarshalAcl(acl),
            pourCoverage: pourCoveragePct, // TODO: only allow setting if originally undefined
            condensationCoverageAtSealTime: condensationCoveragePct,  // TODO: only allow setting if originally undefined
            wetAtCooledTime: wetAtCooledTime,
            agarOnOutsideAtPourTime: agarOnOutsideAtPourTime,
        }

        try {

            // Pics
            const picsInfo = resolvePicsFormData(images)
            dataObj.images = picsInfo.obj
            setFormImages(formData, "newPic", picsInfo.images)
            // Contams
            const contamsInfo = resolveContamsFormData(contams)
            dataObj.contams = contamsInfo.obj
            setFormImages(formData, "newContam", contamsInfo.images)
            // Set data on form
            setFormData(formData, dataObj)
        } catch (caught: any) {
            console.log("error in submit")
            setErr(JSON.stringify(caught))
            return
        }
        console.log("submitting update request")
        DoUpdateMultipartRequest("plate",initial._id, formData, AssertPlate, allCookies(cookies))
            .then(v=>{
                updateInitial(new PlateData(v))
                console.log("updated initial state")
            })
            .catch(e=>{
                setErr("Error in parsing updated plate: "+JSON.stringify(e))
            })
    }
    const ovcs: OnViewCreatorQuadCol[] = [
        WriteRfidOvcArea(initial._id),
        // TODO: fruit?
        // TODO: create spore print
        // TODO: creat spore swab
    ]
    return (
        <DisplayFormWrapper entryType={"plate"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <ID props={{id:data._id, txt:"Plate", entryType:"plate", linkPage:false}}/>
            <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>
            <MostRecentImageDisplay data={initial.mostRecentImage} headerLevel={headerLevel} showHeader={false}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated}
                                                disposed={disposed}
                                                readonly={readonly} setDisposedOnParent={setDisposed}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <SpeciesSubspeciesArea subspecies={initial.subspecies} species={initial.species}/>
                    <AgarBatchArea agarBatchId={initial.agarBatch} headerLevel={headerLevel}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <InnocDisplay innoc={initial.innoc}/>
                    <ParentDisplay parent={initial.parent} parentType={initial.parentType} headerLevel={headerLevel}/>
                    <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable} readonly={readonly}/>
                    <SaleArea sale={sale} setSale={setSale} readonly={readonly} headerLevel={headerLevel}
                              canCreateSale={true}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    {/* TODO: SIZING ON ENTRY FIELDS*/}
                    <TestAndValidate todos={["ensure unset values for pour and condens work as expected"]}>
                    <PourCoverageFieldDisplay pourCoveragePct={pourCoveragePct} updateParent={setPourCoveragePct}/>{/* TODO: STATIC IF PRE-SET*/}
                    {initial.condensationCoverageAtSealTime ? <div>{"Condensation Coverage: "+initial.condensationCoverageAtSealTime+"%"}</div>:
                        <CondensationCoverageFieldDisplay coverage={condensationCoveragePct/* TODO: STATIC IF PRE-SET*/} updateParent={setCondensationCoveragePct}/>}
                    </TestAndValidate>
                    <YesNoSelector pre={"Wet at cooled time: "} initial={initial.wetAtCooledTime}
                                   updateParent={setWetAtCoolTime}/>
                    <YesNoSelector pre={"Agar on outside at pour time: "} initial={initial.agarOnOutsideAtPourTime}
                                   updateParent={setAgarOnOutsideAtPourTime}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            <TransfersOutDisplay headerTxt={"Transfers"} thisId={initial._id} thisEntryType={"plate"}
                                 transfersOut={transfersOut}
                                 allowNewTransferCreation={!readonly}/>{/* TODO: have this rely on dictation as well to figure out if we want it open on the screen*/}
            <PicsDisplay pix={initial.pics || []} readonly={readonly}
                         headerLevel={headerLevel} updateParent={setImages}/>{/* Pics */}
            <ContamsDisplay initial={initial.contamination || []} updateParent={setContams}
                            readonly={readonly} headerLevel={headerLevel}/>
            {/* TODO: SOMEHOW SHOVE THE DICTAPHONE INTO THE NotesFormArea*/}
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay initial={acl} readonly={readonly} updateParent={setAcl}/>
                {/* TODO: when user edits perms, it is adding the user to the perms as a writer. Ensure it does not do this anymore!*/}
                {/* TODO: removing users is not autoupdating in the UI (the removed user is re-added on submit...). Make sure those changes are shown immediately...*/}

            </TogglableAreaWithDepth>

            {readonly || <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}
            {/*<ViewPageDictaphone doUpdate={submit}/>/!* TODO: FIXME *!/*/}
        </DisplayFormWrapper>
    )
}

function PourCoverageFieldDisplay({pourCoveragePct, updateParent}: {
    pourCoveragePct?: number,
    updateParent?: (cov: number) => void
}) {
    const header = "Pour Coverage: "
    if (!pourCoveragePct) {
        return <div>{header + (pourCoveragePct ? pourCoveragePct + "%" : "unknown")}</div>
    }
    const [pourCoverage, setPourCoverage] = useState(pourCoveragePct)
    return <div>{header}<InputNumberWithSmallTitle value={"" + pourCoverage} readonly={false} min={0} max={100} step={1}
                                                   mode={"integer"} onChange={(s) => {
        const temp = Number(s)
        updateParent ? updateParent(temp) : console.warn("pourCoverage has no updateParent prop")
        setPourCoverage(temp)
    }}/>{"%"}</div>
}

function CondensationCoverageFieldDisplay({coverage, updateParent}: {
    coverage?: number,
    updateParent?: (cov: number) => void
}) {
    const header = "Condensation Coverage: "
    if (!coverage) {
        return <div>{header + (coverage ? coverage + "%" : "unknown")}</div>
    }
    const [val, setVal] = useState(coverage)
    return <div>{header}<InputNumberWithSmallTitle value={"" + val} readonly={false} min={0} max={100} step={1}
                                                   mode={"integer"} onChange={(s) => {
        const temp = Number(s)
        updateParent && updateParent(temp)
        setVal(temp)
    }}/>{"%"}</div>
}

// TODO: USE?
export function CondensationCoverageSlider({defaultValue, onChange}:
                                           {
                                               defaultValue: number,
                                               onChange: (event: Event, value: number, activeThumb: number) => void,
                                           }) {
    return <NumberSlider defaultValue={defaultValue} min={0} max={100} label={"Condensation Coverage (%)"}
                         onChange={onChange}/>
}

export function NumberSlider({min, max, label, defaultValue, onChange}:
                             {
                                 min: number,
                                 max: number,
                                 label: string,
                                 defaultValue: number,
                                 onChange: (event: Event, value: number, activeThumb: number) => void,
                             }) {
    return (
        <Box sx={{width: 300}}> {/* TODO: fix box size*/}
            <Slider
                min={min} max={max} defaultValue={defaultValue} step={1}
                size="medium" // small, medium, large
                aria-label={label} // Label

                valueLabelDisplay="on" /* Can be off, on, or auto */
                marks={[{value: 0, label: "none"},
                    {value: 50, label: "half"},
                    {value: 100, label: "complete"},
                ]}
                getAriaValueText={(
                    value: number,
                    index: number,
                ) => {
                    return value.toString()
                }}
                onChange={onChange}
            />
        </Box>
    );
}

function CondensationCoverageSelector({coverage, updateParent}: {
    coverage?: number,
    updateParent?: (cov?: number) => void
}) {
    return <OptionalSliderSelector txt={"Condensation Coverage: "} label={"Condensation Coverage: "} min={0} max={100}
                                   def={50} initial={coverage} updateParent={updateParent}/>
}

function OptionalSliderSelector({txt, label, initial, min, max, updateParent, def}: {
    txt: string,
    label: string,
    initial?: number,
    min: number,
    max: number,
    updateParent?: (cov?: number) => void,
    def: number,
}) {
    const [isDefined, setIsUndefined] = useState(initial !== undefined)
    const [val, setVal] = useState(initial || 50)
    return <div className={"inlineChildren"}>
        <div>{txt}</div>
        {/* TODO: slider stuff should have white text. Slider should disappear when box unchecked */}
        {isDefined && <div className={"ccSelSlider"}>
            <NumberSlider defaultValue={def} min={min} max={max} label={label}
                          onChange={(e, v, a) => {
                              e.stopPropagation();
                              setVal(v)
                              updateParent && updateParent(v)
                          }}/>
        </div>}
        <input className={"ml-[1rem]"} type="checkbox" checked={isDefined} onChange={() => {
            setIsUndefined(!isDefined)
            updateParent && updateParent(undefined)
        }}/>
    </div>
}

export function PlateImportDisplay({}: ImportDisplayInput) {
    const [created, setCreated] = useState<number>(Date.now())
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [subspecies, setSubspecies] = useState<string | undefined>(undefined)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(undefined)
    const [generation, setGeneration] = useState<number | undefined>(undefined)
    const [pourCoverage, setPourCoverage] = useState<number | undefined>(undefined)
    const [imageFile, setImageFile] = useState<File | undefined>(undefined)
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const cookies = useContext(CookiesContext)
    const ImportPlate = () => {
        const formData = new FormData()
        const dataObj: any = {
            creationDate: created,
            // Optionals
            species: species?._id,
            subspecies: subspecies,
            knownFruitable: knownFruitable,
            generation: generation,
            pourCoverage: pourCoverage, // TODO: ensure covered in go
            writeTagTo: writeTagTo,
        }
        if (imageFile !== undefined) {
            formData.set("image", imageFile, "image")
        }
        setFormData(formData, dataObj)
        MultipartImportRequest(formData, "plate", AssertPlate, setErr, allCookies(cookies))
    }
    return <ImportEntryFormWrapper entryType={"plate"}>
        {err != undefined && <div>{"Error: " + err}</div>}
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <ExistingSpeciesSubspeciesSelector doSelectSpecies={setSpecies} doSelectSubspecies={setSubspecies}/>
        {/*<div className={"centerH"}>*/}
        {/*    <ExistingSpeciesSelector initialSpecies={species?._id}*/}
        {/*                             doSelect={(spec?: SpeciesData) => {*/}
        {/*                                 setSpecies(spec)*/}
        {/*                                 setSubspecies(undefined)}}/>*/}
        {/*</div>*/}
        {/*{species !== undefined ? <div className={"centerH"}>*/}
        {/*    <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}/>*/}
        {/*</div> : null}*/}
        <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable}/>
        <GenerationInput updateParent={setGeneration}/>
        {/* TODO: Test coverage slider */}
        <PourCoverageSelectorRequired value={pourCoverage} setPourCoverage={setPourCoverage}/>
        <div className={"centerH"}>
            <ImageSelector updateParent={setImageFile}/>
        </div>

        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"bottomButton"} onClick={ImportPlate}>{"Import Plate"}</button>
    </ImportEntryFormWrapper>
}

export function NewPlateForm(
    {handlers,agarBatchIn}: { handlers: NewEntryInput<PlateData>, agarBatchIn?: AgarBatchData }
) {
    const [agarBatch, setAgarBatch] = useState<AgarBatchData | undefined>(agarBatchIn)
    const [condensationCoverageAtSealTime, setCondensationCoverageAtSealTime] = useState<number | undefined>(undefined)
    const [pourCoverage, setPourCoverage] = useState<number | undefined>(undefined)
    const [wetAtCooledTime, setWetAtCooledTime] = useState<boolean | undefined>(undefined)
    const [agarOnOutsideAtPourTime, setAgarOnOutsideAtPourTime] = useState<boolean | undefined>(undefined)
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>(undefined)

    const cookies = useContext(CookiesContext)
    const createPlate = (e: React.MouseEvent) => {
        e.preventDefault()
        if (agarBatch === undefined) {
            setErr("An agar batch must be selected")
            return
        }
        const body: any = {
            agarBatch: agarBatch._id,
            condensationCoverageAtSealTime: condensationCoverageAtSealTime, // TODO: ensure ok on go side
            pourCoverage: pourCoverage, // TODO: ensure ok on go side
            wetAtCooledTime: wetAtCooledTime, // TODO: ensure ok on go side
            agarOnOutsideAtPourTime: agarOnOutsideAtPourTime, // TODO: ensure ok on go side
            notes: notes,
            writeTagTo: writeTagTo,
        }
        DoCreateRequest("plate", body, AssertPlate, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    const sliderClasses="mt-10 mb-10"//{/* TODO: ensure ok! Change from 10!*/}
    return <NewEntryFormWrapper entryType={"plate"}>
        <ErrorDisplay err={err}/>
        {agarBatchIn === undefined && <AgarBatchSelectorCloseable doSelect={setAgarBatch} allowCreation={true} creatorInPage={true/* TODO: is true ok for both?*/}/>}
        <div className={sliderClasses}>
            <PourCoverageSelector value={pourCoverage} setPourCoverage={setPourCoverage}/>
        </div>
        <div className={sliderClasses}>
            <CondensationCoverageSelector coverage={condensationCoverageAtSealTime}
                                      updateParent={setCondensationCoverageAtSealTime}/>
        </div>
        <YesNoSelector pre={"Wet at cooled time: "} initial={undefined} updateParent={setWetAtCooledTime}
                       className={"inlineChildren"}/>
        <YesNoSelector pre={"Agar on outside at pour time: "} initial={undefined}
                       updateParent={setAgarOnOutsideAtPourTime} className={"inlineChildren"}/>
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"bottomButton greenButton"} onClick={createPlate}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function PlateListPageTable({data, onClick, withLink}: ListPageItems<PlateData>) {
    let cols: ListTableColumn<PlateData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Spec", (v) => v.species || ""),
        NewColumn("Subspec", v => v.subspecies || ""),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: PlateData) => {
            return <EntryLinkWrapper props={{entry:v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new PlateData(v)}}/>
}

export function PlateSelectorTable({data, onClick}: ListPageItems<PlateData>) {
    return <PlateListPageTable data={data} onClick={onClick} withLink={true}/>
}

export function PlateSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: PlateData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: PlateData[]): JSX.Element => {
        return <PlateSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"plate"} entryTypes={"plates"} doSelect={doSelect} asserter={AssertPlate}
                                   table={table}>
        {allowCreate && <NewPlateForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingRecentSelector>
}