'use client'

import React, {SyntheticEvent, useState} from "react";
import {IsValidNote, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import {AllEntries, OnViewCreatorQuadCol, SplitAllEntries} from "@/app/components/formSubcomponents/shared";
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
import GenerationArea from "@/app/components/formSubcomponents/generationInput";
import {
    DisplayInput,
    DisposedSaleContamArea,
    HandleJsonResponse,
    HandleTxtResponse,
    ImportDisplayInput,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    resolveContamsFormData,
    resolvePicsFormData,
    SendMultipartRequest,
    setFormData,
    setFormImages, TwoValuePlusUnknownSelector,
    viewUrlFor, YesNoSelector,
} from "@/app/components/common";
import ReaderWriterSelector from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {redirect} from "next/navigation";
import {
    DisposedDisplay,
    ErrorDisplay,
    GensFormDisplay,
    GensInlineDisplay,
    MostRecentImageDisplay,
    ParentDisplay,
    PicsDisplay,
    SpeciesArea,
    SubspeciesArea
} from "@/app/components/formSubcomponents/commonClient";
import {AgarBatchArea, FlexedArea, FlexedSinglesGroup, NotesFormArea,} from "@/app/components/agarBatchClient";
import {
    ContaminationForm,
    ContamsDisplay,
    InitialContamState,
    InitialNotesState,
    IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import {AgarBatchData, AgarBatchSelector} from "@/app/components/agarBatchServer";
import {BaseExternalUrl} from "@/app/components/Constants";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SaleArea} from "@/app/components/saleClient";
import {ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {OnViewCreatorsQuadColArea} from "@/app/components/pcRunClient";
import {DisplayFormWrapper, ImportEntryFormWrapper, NewEntryFormWrapper} from "./lcRecipeClient";
import {InlineEntry} from "@/app/components/agarRecipeClient";
import {SpeciesSubspeciesArea} from "@/app/components/lcClient";
import {InputNumberWithSmallTitle} from "@/app/components/formSubcomponents/numericInput";

export function AssertPlate(input: any): asserts input is PlateData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Plate assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
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
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Plate assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Plate assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['pics', IsValidPicWithNotesIncoming],
        ['contamination', IsValidContamination],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Plate assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export function PourCoverageSelector({value, setPourCoverage}: {
    value?: number,
    setPourCoverage: (value?: number) => void,
}) {
    return <TestAndValidate todos={["DO THIS POUR COVERAGE AREA!"]}>
        <div>{"Pour coverage (% of all):"}</div>
        <div>
            <TestAndValidate todos={["FIX!"]}>
                {"FIX_ME"/* TODO: THIS!*/}
            </TestAndValidate>
        </div>
    </TestAndValidate>
}

export default function PlateDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }: DisplayInput) {
    // TODO: condensationCoverageAtSealTime?: number
    // TODO: pourCoverage?: number
    // TODO: wetAtCooledTime?: boolean
    // TODO: agarOnOutsideAtPourTime?: boolean
    const [initial, setInitial] = useState(data as PlateData)
    // TODO: only use initial below this!
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
    const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
    // Helper states
    const [transfersOut, setTransfersOut] = useState<string[]>(data.transfersOut || [])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const updateInitial = (updated: PlateData) => {
        setInitial(updated) // TODO: ensure verywhere does this
        setImages(InitialPicsEntries(updated.pics))
        setContams(InitialContamState(updated.contamination))
        setKnownFruitable(updated.knownFruitable)
        setSale(updated.sale)
        setDisposed(updated.disposed)
        setNotes(InitialNotesState(updated.notes))
        setTransfersOut(updated.transfersOut || [])
        setAcl(updated.acl) // TODO: ensure verywhere does this
    }
    const submit = () => {
        console.log("submitting update request")
        let body = new FormData()
        let dataObj: any = {
            knownFruitable: knownFruitable,
            sale: sale,
            disposed: disposed,
            notes: notes,
            writeTagTo: writeTagTo,
            acl: MarshalAcl(acl),
            pourCoverage: pourCoveragePct, // TODO: ensure covered by go
            condensationCoverageAtSealTime: condensationCoveragePct, // TODO: ensure covered by go
            wetAtCooledTime: wetAtCooledTime, // TODO: ensure covered by go
            agarOnOutsideAtPourTime: agarOnOutsideAtPourTime // TODO: ensure covered by go
        }

        try {
            // Pics
            let picsInfo = resolvePicsFormData(images)
            let newImages = picsInfo.images
            dataObj.images = picsInfo.obj
            // Contams
            let contamsInfo = resolveContamsFormData(contams)
            let newContams = contamsInfo.images
            dataObj.contams = contamsInfo.obj
            // Set data on form
            setFormData(body, dataObj)
            setFormImages(body, "newPic", newImages)
            setFormImages(body, "newContam", newContams)
        } catch (caught: any) {
            console.log("error in submit")
            setErr(JSON.stringify(caught))
            return
        }
        SendMultipartRequest(BaseExternalUrl + "/db/update/plate/" + initial._id, cookies, body)
            .then(HandleJsonResponse)
            .then((entry) => { // TODO: ensure plate update comes back valid!
                AssertPlate(entry)
                console.log("Attempting to update initial")
                updateInitial(entry) // TODO: validate working properly
            }).catch((er) => {
            console.log("error in getting response: " + JSON.stringify(er))
            setErr(JSON.stringify(er))
        });
    }
    const ovcs: OnViewCreatorQuadCol[] = [
        // TODO: anything here?
    ] // TODO: THIS!
    return (
        <DisplayFormWrapper entryType={"plate"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <ID txt={"Plate"} id={initial._id} entryType={"plate"} linkPage={false}/>
            <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
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
                    <PourCoverageFieldDisplay pourCoveragePct={pourCoveragePct} updateParent={setPourCoveragePct}/>
                    <CondensationCoverageFieldDisplay coverage={condensationCoveragePct}
                                                      updateParent={setCondensationCoveragePct}/>
                    <YesNoSelector pre={"Wet at cooled time: "} initial={initial.wetAtCooledTime}
                                   updateParent={setWetAtCoolTime}/>
                    <YesNoSelector pre={"Agar on outside at pour time: "} initial={initial.agarOnOutsideAtPourTime}
                                   updateParent={setAgarOnOutsideAtPourTime}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            <TransfersOutDisplay headerTxt={"Transfers"} thisId={initial._id} thisEntryType={"plate"}
                                 transfersOut={transfersOut}
                                 allowNewTransferCreation={!readonly} cookies={cookies}/>
            <PicsDisplay pix={images} readonly={readonly}
                         headerLevel={headerLevel} updateParent={setImages}/>{/* Pics */}
            <ContamsDisplay initial={initial.contamination || []} current={contams} updateParent={setContams}
                            readonly={readonly} headerLevel={headerLevel}/>
            {/* TODO: REDO THE NOTESFORMAREA and NotesArea???*/}
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
            </TogglableAreaWithDepth>

            {readonly || <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>}
            {readonly || <button className={"bottomButton greenButton"} onClick={(e)=>{
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}
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
        updateParent && updateParent(temp) // TODO: ensure ok
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
        updateParent && updateParent(temp) // TODO: ensure ok
        setVal(temp)
    }}/>{"%"}</div>
}

export function PlateImportDisplay({cookies}: ImportDisplayInput) {
    // TODO: condensationCoverageAtSealTime?: number
    // TODO: pourCoverage?: number
    // TODO: wetAtCooledTime?: boolean
    // TODO: agarOnOutsideAtPourTime?: boolean
    const [created, setCreated] = useState<number>(Date.now())
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>(undefined)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(undefined)
    const [generation, setGeneration] = useState<number | undefined>(undefined)
    const [pourCoverage, setPourCoverage] = useState<number | undefined>(undefined)
    const [imageFile, setImageFile] = useState<File | undefined>(undefined)
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const ImportPlate = () => {
        if (species === undefined) {
            setErr("Species must be set!")
            return
        }
        let formData = new FormData()
        let dataObj: any = {
            created: created,
            species: species._id,
            //perms: perms,
            // Optionals
            subspecies: subspecies?._id,
            knownFruitable: knownFruitable,
            generation: generation,
            pourCoverage: pourCoverage,
        }
        if (imageFile !== undefined) {
            formData.set("image", imageFile, "image")
        }
        writeTagTo && (dataObj.writeTagTo = writeTagTo)
        setFormData(formData, dataObj)
        SendMultipartRequest(BaseExternalUrl + "/db/import/plate", cookies, formData)
            .then(HandleTxtResponse)
            .then((newId) => {
                redirect(BaseExternalUrl + "/view/plate/" + newId)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    return <ImportEntryFormWrapper entryType={"plate"}>
        {err != undefined && <div>{"Error: " + err}</div>}
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <div className={"centerH"}>
            <ExistingSpeciesSelector initialSpecies={species?._id}
                                     doSelect={(spec?: SpeciesData) => {
                                         setSpecies(spec)
                                         setSubspecies(undefined)
                                     }/*cookies={cookies}*/}/>
        </div>
        {species !== undefined ? <div className={"centerH"}>
            <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies/*cookies={cookies}*/}/>
        </div> : null}
        <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable}/>
        <GenerationArea initial={generation} readonly={false} updateParent={setGeneration}/>
        <div
            className={"centerH"}> {/* TODO: SOMETHING IS GOING WEIRD HERE WITH SIZING DIVS! BOTTOM DIV STICKS OUT WITH FILECHOOSE BUTTON */}
            <ImageSelector updateParent={setImageFile}/>
        </div>
        <PourCoverageSelector value={pourCoverage} setPourCoverage={setPourCoverage}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"bottomButton"} onClick={ImportPlate}>{"Import Plate"}</button>
    </ImportEntryFormWrapper>
}

export function CreatedUpdatedDisposedArea( // TODO: MOVE
    {created, updated, disposed, readonly, setDisposedOnParent}: {
        created: number,
        updated: number,
        disposed?: number,
        readonly: boolean,
        setDisposedOnParent?: (n?: number) => void,
    }
) {
    return <>
        {/*<div className={"createdUpdatedDisposedArea"}>*/}
        <DateArea pre={"Created: "} when={created} readonly={true}/>
        <DateArea pre={"Updated: "} when={updated} readonly={true}/>
        <DisposedDisplay readonly={readonly} disposed={disposed} setDisposedOnParent={setDisposedOnParent}/>
        {/*</div>*/}
    </>
}

export function NewPlateForm(
    {handlers}: { handlers: NewEntryInput<PlateData>, agarBatchIn?: AgarBatchData }
) {
    // TODO: condensationCoverageAtSealTime?: number
    // TODO: pourCoverage?: number
    // TODO: wetAtCooledTime?: boolean
    // TODO: agarOnOutsideAtPourTime?: boolean
    const [agarBatch, setAgarBatch] = useState<AgarBatchData | undefined>(undefined)
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    // TODO: handle isTopLevel
    const createPlate = (e: React.MouseEvent) => {
        e.preventDefault()
        if (agarBatch === undefined) {
            setErr("An agar batch must be selected")
            return
        }
        window.location.assign(viewUrlFor("plate", "testPlate")) // TODO: DELETE!
        return // TODO: DELETE!
        let body: any = {agarBatch: agarBatch}
        writeTagTo && (body.writeTagTo = writeTagTo)
        fetch(BaseExternalUrl + "/create/plate", { // TODO: ensure correct
            method: "POST",
            headers: {
                credentials: 'include',
                // TODO: may need 'Cookie': cookies,
                'Content-type': 'application/json'
                // TODO: THIS!
            },
            body: JSON.stringify({
                agarBatch: agarBatch,
                writeTagTo: writeTagTo,
            })
        })
            .then(HandleJsonResponse)
            .then((entry) => {
                AssertPlate(entry)
                handlers.onCreate && handlers.onCreate(entry)
            })
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
    }
    return <NewEntryFormWrapper entryType={"plate"}>
        <ErrorDisplay err={err}/>
        <AgarBatchSelector doSelect={setAgarBatch} allowCreation={true} creatorInPage={true}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"bottomButton"} onClick={createPlate}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function PlateInline({data, expandByDefault, onClick, showMainPageButton, idIsLink}: InlineProps<PlateData>) {
    const [expanded, setExpanded] = useState(expandByDefault)
    const b58id = data._id
    // TODO: condensationCoverageAtSealTime?: number
    // TODO: pourCoverage?: number
    // TODO: wetAtCooledTime?: boolean
    // TODO: agarOnOutsideAtPourTime?: boolean
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={b58id} txt={"Plate"} entryType={"plate"} allowOpenMainPage={showMainPageButton}
                linkPage={idIsLink}/>
            <SpeciesArea readonly={true} initial={data.species}/>
            <SubspeciesArea readonly={true} currentSpecies={data.species} initialSub={data.subspecies}/>
            <KnownFruitableArea initial={data.knownFruitable} readonly={true}/>
            <GensInlineDisplay gensSinceSpore={data.genSpore} gensSinceFruitOrSpore={data.genFruitOrSpore}/>
            <DisposedSaleContamArea sale={data.sale} disposed={data.disposed} contams={data.contamination}/>
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            <AgarBatchArea agarBatchId={data.agarBatch} offset={-1}/>
            {data.condensationCoverageAtSealTime !== undefined ? <div>
                {"Initial condensation coverage: " + data.condensationCoverageAtSealTime + "%"}
            </div> : <div>
                {"Initial condensation coverage: none or unknown"}
            </div>}
            {data.pourCoverage !== undefined ? <div>
                {"Pour coverage: " + data.pourCoverage + "%"}
            </div> : <div>
                {"Pour coverage: none or unknown"}
            </div>}
            {data.wetAtCooledTime !== undefined ? <div>
                {"Initial wetness: " + (data.wetAtCooledTime ? "wet" : "perfect")}
            </div> : <div>
                {"Initial wetness: unknown"}
            </div>}
            {data.agarOnOutsideAtPourTime !== undefined ? <div>
                {"Agar on outside when poured: " + (data.agarOnOutsideAtPourTime ? "yes" : "no")}
            </div> : <div>
                {"Agar on outside when poured: not likely, unknown"}
            </div>}
            {/*TODO: <ProjectsArea allowCreate={false} projects={data.perms?.projectPerms.ids} readonly={true} headerLevel={headerLevel}*/}
            {/*              offset={-1} allowRemove={false}/>*/}
            <NotesAreaInline notes={data.notes} offset={-1}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                                                     expanded={expanded}/>
    </InlineEntry>
}

// export function PlateListDisplay({data, onClick}: SingleListProps<PlateData>) {
//     return <div>
//         {data.map((b, i) => {
//             return <PlateInline data={b} onClick={() => {
//                 onClick(b)
//             }} key={i}/>
//         })}
//     </div>
// }